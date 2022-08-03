package impl

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	logger "github.com/rs/zerolog/log"
	"github.com/textileio/go-tableland/internal/tableland"
	"github.com/textileio/go-tableland/pkg/eventprocessor"
	"github.com/textileio/go-tableland/pkg/eventprocessor/eventfeed"
	"github.com/textileio/go-tableland/pkg/eventprocessor/impl/executor"
	"github.com/textileio/go-tableland/pkg/parsing"
)

type blockScope struct {
	txn    *sql.Tx
	log    zerolog.Logger
	parser parsing.SQLValidator
	acl    tableland.ACL

	scopeVars scopeVars
}

type scopeVars struct {
	ChainID          tableland.ChainID
	MaxTableRowCount int
}

func newBlockScope(
	txn *sql.Tx,
	scopeVars scopeVars,
	parser parsing.SQLValidator,
	acl tableland.ACL,
	blockNum int64,
	blockHash string,
) *blockScope {
	log := logger.With().
		Str("component", "blockscope").
		// TODO(jsign): fix all with chain_id
		Int64("chainID", int64(scopeVars.ChainID)).
		Int64("block_number", blockNum).
		Str("blockHash", blockHash).
		Logger()

	return &blockScope{
		txn:       txn,
		log:       log,
		parser:    parser,
		acl:       acl,
		scopeVars: scopeVars,
	}
}

// ExecuteEvents executes all the events in a txn atomically.
//
// If the events execution is successful, it returns the result.
// If the events execution isn't successful, we have two cases:
// 1) If caused by controlled error (e.g: invalid SQL), it will return a (res, nil) where
//    res contains the error message.
// 2) If caused by uncontrolled error (e.g: can't access the DB), it returns ({}, err). The caller should retry
//    executing this transaction events when the underlying infrastructure problem is solved.
func (bs *blockScope) ExecuteTxnEvents(ctx context.Context, evmTxn eventfeed.TxnEvents) (executor.TxnExecutionResult, error) {
	// Create nested transaction from the blockScope. All the events for this trasaction will be executed here.
	if _, err := bs.txn.ExecContext(ctx, "SAVEPOINT txnscope"); err != nil {
		return executor.TxnExecutionResult{}, fmt.Errorf("creating savepoint: %s", err)
	}

	ts := &txnScope{
		parser:    bs.parser,
		acl:       bs.acl,
		scopeVars: bs.scopeVars,
		log: logger.With().
			Str("component", "txnscope").
			Int64("chainID", int64(bs.scopeVars.ChainID)).
			Str("txnHash", evmTxn.TxnHash.String()).
			Logger(),

		txn: bs.txn,
	}
	tableID, err := ts.executeTxnEvents(ctx, bs.txn, evmTxn)
	if err != nil {
		if _, err := bs.txn.ExecContext(ctx, "ROLLBACK TO txnscope"); err != nil {
			return executor.TxnExecutionResult{}, fmt.Errorf("rollbacking savepoint: %s", err)
		}
		return executor.TxnExecutionResult{}, fmt.Errorf("executing query: %w", err)
	}
	if _, err := bs.txn.ExecContext(ctx, "RELEASE SAVEPOINT txnscope"); err != nil {
		return executor.TxnExecutionResult{}, fmt.Errorf("releasing savepoint: %s", err)
	}

	return tableID, nil
}

func (bs *blockScope) GetLastProcessedHeight(ctx context.Context) (int64, error) {
	r := bs.txn.QueryRowContext(
		ctx,
		"SELECT block_number FROM system_txn_processor WHERE chain_id=?1 LIMIT 1",
		bs.scopeVars.ChainID)
	var blockNumber int64
	if err := r.Scan(&blockNumber); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("get last block number query: %s", err)
	}
	return blockNumber, nil
}

// TODO(jsign): move to Commit?
func (bs *blockScope) SetLastProcessedHeight(ctx context.Context, height int64) error {
	tag, err := bs.txn.ExecContext(
		ctx,
		"UPDATE system_txn_processor SET block_number=?1 WHERE chain_id=?2",
		height, bs.scopeVars.ChainID)
	if err != nil {
		return fmt.Errorf("update last processed block number: %s", err)
	}
	ra, err := tag.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %s", err)
	}
	if ra != 1 {
		if _, err := bs.txn.ExecContext(ctx,
			"INSERT INTO system_txn_processor (block_number, chain_id) VALUES (?1, ?2)",
			height,
			bs.scopeVars.ChainID,
		); err != nil {
			return fmt.Errorf("inserting first processed height: %s", err)
		}
	}
	return nil
}

func (bs *blockScope) SaveTxnReceipts(ctx context.Context, rs []eventprocessor.Receipt) error {
	for _, r := range rs {
		tableID := sql.NullInt64{Valid: false}
		if r.TableID != nil {
			tableID.Valid = true
			tableID.Int64 = r.TableID.ToBigInt().Int64()
		}
		if r.Error != nil {
			*r.Error = strings.ToValidUTF8(*r.Error, "")
		}
		if _, err := bs.txn.ExecContext(
			ctx,
			`INSERT INTO system_txn_receipts (chain_id,txn_hash,error,table_id,block_number,index_in_block) 
				 VALUES (?1,?2,?3,?4,?5,?6)`,
			r.ChainID, r.TxnHash, r.Error, tableID, r.BlockNumber, r.IndexInBlock); err != nil {
			return fmt.Errorf("insert txn receipt: %s", err)
		}
	}
	return nil
}

func (bs *blockScope) TxnReceiptExists(ctx context.Context, txnHash common.Hash) (bool, error) {
	r := bs.txn.QueryRowContext(
		ctx,
		`SELECT 1 from system_txn_receipts WHERE chain_id=?1 and txn_hash=?2`,
		bs.scopeVars.ChainID, txnHash.Hex())
	var dummy int
	err := r.Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get txn receipt: %s", err)
	}
	return true, nil
}

// Close closes gracefully the block scope.
// Clients should *always* `defer Close()` when opening batches.
func (bs *blockScope) Close() error {
	// Calling rollback is always safe:
	// - If Commit() wasn't called, the result is a rollback.
	// - If Commit() was called, *sql.Txn guarantees is a noop.
	if err := bs.txn.Rollback(); err != nil {
		if err != sql.ErrTxDone {
			return fmt.Errorf("closing batch: %s", err)
		}
	}
	return nil
}

func (b *blockScope) Commit() error {
	if err := b.txn.Commit(); err != nil {
		return fmt.Errorf("commit txn: %s", err)
	}
	return nil
}
