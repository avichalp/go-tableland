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
	"github.com/textileio/go-tableland/pkg/dbhash"
	"github.com/textileio/go-tableland/pkg/eventprocessor"
	"github.com/textileio/go-tableland/pkg/eventprocessor/eventfeed"
	"github.com/textileio/go-tableland/pkg/eventprocessor/impl/executor"
	"github.com/textileio/go-tableland/pkg/parsing"
	"golang.org/x/crypto/sha3"
)

type blockScope struct {
	txn    *sql.Tx
	log    zerolog.Logger
	parser parsing.SQLValidator
	acl    tableland.ACL

	scopeVars scopeVars

	closed func()
}

type scopeVars struct {
	ChainID          tableland.ChainID
	MaxTableRowCount int
	BlockNumber      int64
}

func newBlockScope(
	txn *sql.Tx,
	scopeVars scopeVars,
	parser parsing.SQLValidator,
	acl tableland.ACL,
	closed func(),
) *blockScope {
	log := logger.With().
		Str("component", "blockscope").
		Int64("chain_id", int64(scopeVars.ChainID)).
		Int64("block_number", scopeVars.BlockNumber).
		Logger()

	return &blockScope{
		txn:       txn,
		log:       log,
		parser:    parser,
		acl:       acl,
		scopeVars: scopeVars,
		closed:    closed,
	}
}

// ExecuteEvents executes all the events in a txn atomically.
//
// If the events execution is successful, it returns the result.
// If the events execution isn't successful, we have two cases:
//  1. If caused by controlled error (e.g: invalid SQL), it will return a (res, nil) where
//     res contains the error message.
//  2. If caused by uncontrolled error (e.g: can't access the DB), it returns ({}, err). The caller should retry
//     executing this transaction events when the underlying infrastructure problem is solved.
func (bs *blockScope) ExecuteTxnEvents(
	ctx context.Context,
	evmTxn eventfeed.TxnEvents,
) (executor.TxnExecutionResult, error) {
	// Create nested transaction from the blockScope. All the events for this transaction will be executed here.
	if _, err := bs.txn.ExecContext(ctx, "SAVEPOINT txnscope"); err != nil {
		return executor.TxnExecutionResult{}, fmt.Errorf("creating savepoint: %s", err)
	}

	ts := &txnScope{
		scopeVars: bs.scopeVars,

		parser:            bs.parser,
		statementResolver: newWriteStatementResolver(evmTxn.TxnHash.Hex(), bs.scopeVars.BlockNumber),

		acl: bs.acl,

		log: logger.With().
			Str("component", "txnscope").
			Int64("chain_id", int64(bs.scopeVars.ChainID)).
			Str("txn_hash", evmTxn.TxnHash.String()).
			Logger(),

		txn: bs.txn,
	}
	res, err := ts.executeTxnEvents(ctx, evmTxn)
	if err != nil || res.Error != nil {
		if _, err := bs.txn.ExecContext(ctx, "ROLLBACK TO txnscope"); err != nil {
			return executor.TxnExecutionResult{}, fmt.Errorf("rollbacking savepoint: %s", err)
		}
	}
	if err != nil {
		return executor.TxnExecutionResult{}, fmt.Errorf("executing query: %w", err)
	}
	if _, err := bs.txn.ExecContext(ctx, "RELEASE SAVEPOINT txnscope"); err != nil {
		return executor.TxnExecutionResult{}, fmt.Errorf("releasing savepoint: %s", err)
	}

	return res, nil
}

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
			`INSERT INTO system_txn_receipts (chain_id,txn_hash,error,error_event_idx,table_id,block_number,index_in_block) 
				 VALUES (?1,?2,?3,?4,?5,?6,?7)`,
			r.ChainID, r.TxnHash, r.Error, r.ErrorEventIdx, tableID, r.BlockNumber, r.IndexInBlock); err != nil {
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

func (bs *blockScope) StateHash(ctx context.Context, chainID tableland.ChainID) (executor.StateHash, error) {
	hash, err := dbhash.DatabaseStateHash(ctx, bs.txn, []dbhash.Option{
		dbhash.WithFetchSchemasQuery(
			fmt.Sprintf(`SELECT tbl_name, sql 
				FROM sqlite_schema
			    WHERE name NOT LIKE 'sqlite_%%'  
				AND name LIKE '%%\_%d\_%%' ESCAPE '\'
				AND type = 'table'
				UNION ALL
				SELECT tbl_name, sql 
				FROM sqlite_schema
				WHERE name in ('registry', 'system_acl', 'system_controller', 'system_txn_receipts')
				ORDER BY tbl_name;`, chainID),
		),
		dbhash.WithPerTableQueryFn(func(tableName string) string {
			switch tableName {
			case "registry":
				return fmt.Sprintf(`SELECT id, chain_id, controller, prefix, structure 
							FROM registry 
							WHERE chain_id = %d 
							ORDER BY id`, chainID)
			case "system_acl":
				return fmt.Sprintf(`SELECT chain_id, table_id, controller, privileges 
							FROM system_acl 
							WHERE chain_id = %d 
							ORDER BY table_id`, chainID)
			case "system_controller":
				return fmt.Sprintf(`SELECT chain_id, table_id, controller 
							FROM system_controller 
							WHERE chain_id = %d
							ORDER BY table_id`, chainID)
			case "system_txn_receipts":
				return fmt.Sprintf(`SELECT chain_id, block_number, index_in_block, txn_hash, error, table_id 
							FROM system_txn_receipts 
							WHERE chain_id = %d 
							ORDER BY table_id, block_number, index_in_block`, chainID)
			default:
				return fmt.Sprintf("SELECT * FROM %s ORDER BY rowid", tableName)
			}
		}),
	}...)
	if err != nil {
		return executor.StateHash{}, fmt.Errorf("database state hash: %s", err)
	}

	return executor.NewStateHash(chainID, bs.scopeVars.BlockNumber, hash), nil
}

func (bs *blockScope) CalculateTreeLeaves(ctx context.Context, chainID tableland.ChainID) error {
	rows, err := bs.txn.QueryContext(ctx, "select prefix, id from registry where chain_id = ?1 ORDER BY rowid", chainID)
	if err != nil {
		return fmt.Errorf("fetching tables from registry: %s", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var tablePrefix string
		var tableID int
		if err := rows.Scan(&tablePrefix, &tableID); err != nil {
			return fmt.Errorf("scanning table name: %s", err)
		}

		if err := bs.calculateTreeLeavesForTable(ctx, chainID, tablePrefix, tableID); err != nil {
			return fmt.Errorf("calculate leaves for table: %s", err)
		}
	}

	return nil
}

func (bs *blockScope) calculateTreeLeavesForTable(
	ctx context.Context,
	chainID tableland.ChainID,
	tablePrefix string,
	tableID int,
) error {
	tableName := fmt.Sprintf("%s_%d_%d", tablePrefix, chainID, tableID)
	tableRows, err := bs.txn.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return fmt.Errorf("fetching rows from %s: %s", tableName, err)
	}
	defer func() {
		_ = tableRows.Close()
	}()

	cols, err := tableRows.Columns()
	if err != nil {
		return fmt.Errorf("getting the columns of row: %s", err)
	}

	rawBuffer := make([]sql.RawBytes, len(cols))
	scanCallArgs := make([]interface{}, len(rawBuffer))
	for i := range rawBuffer {
		scanCallArgs[i] = &rawBuffer[i]
	}

	leaves := []byte{}
	for tableRows.Next() {
		if err := tableRows.Scan(scanCallArgs...); err != nil {
			return fmt.Errorf("table row scan: %s", err)
		}

		rowHash := sha3.New256()
		for _, col := range rawBuffer {
			rowHash.Write(col)
		}

		leaves = append(leaves, rowHash.Sum(nil)...)
	}

	if _, err := bs.txn.ExecContext(ctx,
		"INSERT INTO system_tree_leaves (prefix, chain_id, table_id, block_number, leaves) VALUES (?1, ?2, ?3, ?4, ?5)",
		tablePrefix,
		bs.scopeVars.ChainID,
		tableID,
		bs.scopeVars.BlockNumber,
		leaves,
	); err != nil {
		return fmt.Errorf("inserting leaves %s: %s", tableName, err)
	}

	return nil
}

// Close closes gracefully the block scope.
// Clients should *always* `defer Close()` when opening block scopes.
func (bs *blockScope) Close() error {
	defer bs.closed()

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

// Commit confirms all successful transaction processing executed in the block scope.
func (bs *blockScope) Commit() error {
	if err := bs.txn.Commit(); err != nil {
		return fmt.Errorf("commit db txn: %s", err)
	}
	return nil
}

type writeStatmentResolver struct {
	txnHash     string
	blockNumber int64
}

func newWriteStatementResolver(txnHash string, blockNumber int64) *writeStatmentResolver {
	return &writeStatmentResolver{txnHash: txnHash, blockNumber: blockNumber}
}

func (wqr *writeStatmentResolver) GetTxnHash() string {
	return wqr.txnHash
}

func (wqr *writeStatmentResolver) GetBlockNumber() int64 {
	return wqr.blockNumber
}
