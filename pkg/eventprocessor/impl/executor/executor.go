package executor

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/textileio/go-tableland/internal/tableland"
	"github.com/textileio/go-tableland/pkg/eventprocessor"
	"github.com/textileio/go-tableland/pkg/eventprocessor/eventfeed"
)

// Executor provides a safe way of executing events contained in an EVM blockchain block.
type Executor interface {
	// NewBlockScope returns a new block scope which can execute events generated by EVM-transactions.
	NewBlockScope(context.Context, int64) (BlockScope, error)

	// Close gracefully closes the executor, waiting for any block scope to be gracefully closed or force closing
	// if the provided context gets cancelled.
	Close(context.Context) error
}

// BlockScope provides a sandbox to execute events generated by each EVM transaction in the block.
// It provides an all or nothing execution at the block level, while allowing each transaction processing to also be
// an all or nothing execution of all the events contained in that transaction.
type BlockScope interface {
	// ExecuteTxnEvents executes atomically all the events in an EVM-transaction, returning the TableID where
	// changes were applied. Changes aren't fully commited to the database until Commit(...) is called.
	// If the execution of events in the transaction fails, the client should distinguish between errors of type
	// ErrQueryExecution which aren't recoverable, and infrastructure errors which are recoverable.
	ExecuteTxnEvents(context.Context, eventfeed.TxnEvents) (TxnExecutionResult, error)

	// GetLastProcessedHeight returns the last processed height.
	GetLastProcessedHeight(ctx context.Context) (int64, error)

	// SetLastprocessedHeight sets a new processed height.
	SetLastProcessedHeight(ctx context.Context, height int64) error

	// SaveTxnReceipts saves a set of transaction receipts.
	SaveTxnReceipts(ctx context.Context, rs []eventprocessor.Receipt) error

	// TxnReceiptExists return true if the provided transaction hash was already processed, and false otherwise.
	TxnReceiptExists(ctx context.Context, txnHash common.Hash) (bool, error)

	// Commit commits all the changes that happened in  previously successful ExecuteTxnEvents(...) calls.
	Commit() error

	// Close gracefully closes the block scope. If Commit(...) called before, it's a noop. If Commit(...) wasn't called,
	// then it will rollback any changes done in previous ExecuteTxnEvents(...) calls.
	Close() error
}

type TxnExecutionResult struct {
	TableID *tableland.TableID
	Error   *string
}

// TODO(jsign): remove from here.
// ErrQueryExecution is an error returned when the query execution failed
// with a cause related to th query itself. Retrying the execution of this query
// will always return an error (e.g: inserting a string in an integer column).
// A query execution failure due to the database being down or any other infrastructure
// problem isn't an ErrQueryExecution error.
type ErrQueryExecution struct {
	Code string
	Msg  string
}

// Error returns a string representation of the query execution error.
func (e *ErrQueryExecution) Error() string {
	return fmt.Sprintf("query execution failed with code %s: %s", e.Code, e.Msg)
}
