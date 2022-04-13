package impl

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/textileio/go-tableland/internal/tableland"
	"github.com/textileio/go-tableland/pkg/nonce"
	"github.com/textileio/go-tableland/pkg/parsing"
	"github.com/textileio/go-tableland/pkg/sqlstore"
)

// ThrottledSQLStorePGX implements a throttled SQLStore interface using pgx.
type ThrottledSQLStorePGX struct {
	store sqlstore.SQLStore
	delay time.Duration
}

// NewThrottledSQLStorePGX creates a new pgx pool and instantiate both the user and system stores.
func NewThrottledSQLStorePGX(store sqlstore.SQLStore, delay time.Duration) sqlstore.SQLStore {
	return &ThrottledSQLStorePGX{store, delay}
}

// GetTable fetchs a table from its UUID.
func (s *ThrottledSQLStorePGX) GetTable(ctx context.Context, id tableland.TableID) (sqlstore.Table, error) {
	return s.store.GetTable(ctx, id)
}

// GetTablesByController fetchs a table from controller address.
func (s *ThrottledSQLStorePGX) GetTablesByController(ctx context.Context,
	controller string) ([]sqlstore.Table, error) {
	return s.store.GetTablesByController(ctx, controller)
}

// Authorize grants the provided address permission to use the system.
func (s *ThrottledSQLStorePGX) Authorize(ctx context.Context, address string) error {
	return s.store.Authorize(ctx, address)
}

// Revoke removes permission to use the system from the provided address.
func (s *ThrottledSQLStorePGX) Revoke(ctx context.Context, address string) error {
	return s.store.Revoke(ctx, address)
}

// IsAuthorized checks if the provided address has permission to use the system.
func (s *ThrottledSQLStorePGX) IsAuthorized(
	ctx context.Context,
	address string,
) (sqlstore.IsAuthorizedResult, error) {
	return s.store.IsAuthorized(ctx, address)
}

// GetAuthorizationRecord gets the authorization record for the provided address.
func (s *ThrottledSQLStorePGX) GetAuthorizationRecord(
	ctx context.Context,
	address string,
) (sqlstore.AuthorizationRecord, error) {
	return s.store.GetAuthorizationRecord(ctx, address)
}

// ListAuthorized returns a list of all authorization records.
func (s *ThrottledSQLStorePGX) ListAuthorized(ctx context.Context) ([]sqlstore.AuthorizationRecord, error) {
	return s.store.ListAuthorized(ctx)
}

// IncrementCreateTableCount increments the counter.
func (s *ThrottledSQLStorePGX) IncrementCreateTableCount(ctx context.Context, address string) error {
	return s.store.IncrementCreateTableCount(ctx, address)
}

// IncrementRunSQLCount increments the counter.
func (s *ThrottledSQLStorePGX) IncrementRunSQLCount(ctx context.Context, address string) error {
	return s.store.IncrementRunSQLCount(ctx, address)
}

// GetACLOnTableByController increments the counter.
func (s *ThrottledSQLStorePGX) GetACLOnTableByController(
	ctx context.Context,
	table tableland.TableID,
	address string) (sqlstore.SystemACL, error) {
	return s.store.GetACLOnTableByController(ctx, table, address)
}

// Read executes a read statement on the db.
func (s *ThrottledSQLStorePGX) Read(ctx context.Context, stmt parsing.SugaredReadStmt) (interface{}, error) {
	data, err := s.store.Read(ctx, stmt)
	time.Sleep(s.delay)

	return data, err
}

// GetNonce returns the nonce stored in the database by a given address.
func (s *ThrottledSQLStorePGX) GetNonce(
	ctx context.Context,
	network string,
	addr common.Address) (nonce.Nonce, error) {
	return s.store.GetNonce(ctx, network, addr)
}

// UpsertNonce updates a nonce.
func (s *ThrottledSQLStorePGX) UpsertNonce(
	ctx context.Context,
	network string,
	addr common.Address,
	nonce int64) error {
	return s.store.UpsertNonce(ctx, network, addr, nonce)
}

// ListPendingTx lists all pendings txs.
func (s *ThrottledSQLStorePGX) ListPendingTx(
	ctx context.Context,
	network string,
	addr common.Address) ([]nonce.PendingTx, error) {
	return s.store.ListPendingTx(ctx, network, addr)
}

// InsertPendingTx insert a new pending tx.
func (s *ThrottledSQLStorePGX) InsertPendingTx(
	ctx context.Context,
	network string,
	addr common.Address,
	nonce int64,
	hash common.Hash) error {
	return s.store.InsertPendingTx(ctx, network, addr, nonce, hash)
}

// DeletePendingTxByHash deletes a pending tx.
func (s *ThrottledSQLStorePGX) DeletePendingTxByHash(ctx context.Context, hash common.Hash) error {
	return s.store.DeletePendingTxByHash(ctx, hash)
}

// Close closes the connection pool.
func (s *ThrottledSQLStorePGX) Close() {
	s.store.Close()
}

// WithTx returns a copy of the current ThrottledSQLStorePGX with a tx attached.
func (s *ThrottledSQLStorePGX) WithTx(tx pgx.Tx) sqlstore.SystemStore {
	return s.store.WithTx(tx)
}
