package impl

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/textileio/go-tableland/pkg/sqlstore"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

// InstrumentedSQLStorePGX implements a instrumented SQLStore interface using pgx.
type InstrumentedSQLStorePGX struct {
	store            sqlstore.SQLStore
	callCount        metric.Int64Counter
	latencyHistogram metric.Int64Histogram
}

// NewInstrumentedSQLStorePGX creates a new pgx pool and instantiate both the user and system stores.
func NewInstrumentedSQLStorePGX(store sqlstore.SQLStore) sqlstore.SQLStore {
	meter := metric.Must(global.Meter("tableland"))
	callCount := meter.NewInt64Counter("tableland.sqlstore.call.count")
	latencyHistogram := meter.NewInt64Histogram("tableland.sqlstore.call.latency")

	return &InstrumentedSQLStorePGX{store, callCount, latencyHistogram}
}

// InsertTable inserts a new system-wide table.
func (s *InstrumentedSQLStorePGX) InsertTable(ctx context.Context, uuid uuid.UUID, controller string) error {
	start := time.Now()
	err := s.store.InsertTable(ctx, uuid, controller)
	latency := time.Since(start).Milliseconds()

	s.callCount.Add(ctx,
		1,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("InsertTable")},
		attribute.KeyValue{Key: "uuid", Value: attribute.StringValue(uuid.String())},
		attribute.KeyValue{Key: "controller", Value: attribute.StringValue(controller)},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)})

	s.latencyHistogram.Record(ctx,
		latency,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("InsertTable")},
		attribute.KeyValue{Key: "uuid", Value: attribute.StringValue(uuid.String())},
		attribute.KeyValue{Key: "controller", Value: attribute.StringValue(controller)},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)},
	)

	return err
}

// GetTable fetchs a table from its UUID.
func (s *InstrumentedSQLStorePGX) GetTable(ctx context.Context, uuid uuid.UUID) (sqlstore.Table, error) {
	start := time.Now()
	table, err := s.store.GetTable(ctx, uuid)
	latency := time.Since(start).Milliseconds()

	s.callCount.Add(ctx,
		1,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("GetTable")},
		attribute.KeyValue{Key: "uuid", Value: attribute.StringValue(uuid.String())},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)})

	s.latencyHistogram.Record(ctx,
		latency,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("GetTable")},
		attribute.KeyValue{Key: "uuid", Value: attribute.StringValue(uuid.String())},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)},
	)

	return table, err
}

// GetTablesByController fetchs a table from controller address.
func (s *InstrumentedSQLStorePGX) GetTablesByController(ctx context.Context,
	controller string) ([]sqlstore.Table, error) {
	start := time.Now()
	tables, err := s.store.GetTablesByController(ctx, controller)
	latency := time.Since(start).Milliseconds()

	s.callCount.Add(ctx,
		1,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("GetTablesByController")},
		attribute.KeyValue{Key: "controller", Value: attribute.StringValue(controller)},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)})

	s.latencyHistogram.Record(ctx,
		latency,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("GetTablesByController")},
		attribute.KeyValue{Key: "controller", Value: attribute.StringValue(controller)},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)},
	)

	return tables, err
}

// Write executes a write statement on the db.
func (s *InstrumentedSQLStorePGX) Write(ctx context.Context, statement string) error {
	start := time.Now()
	err := s.store.Write(ctx, statement)
	latency := time.Since(start).Milliseconds()

	s.callCount.Add(ctx,
		1,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("Write")},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)})

	s.latencyHistogram.Record(ctx,
		latency,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("Write")},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)},
	)

	return err
}

// Read executes a read statement on the db.
func (s *InstrumentedSQLStorePGX) Read(ctx context.Context, statement string) (interface{}, error) {
	start := time.Now()
	data, err := s.store.Read(ctx, statement)
	latency := time.Since(start).Milliseconds()

	s.callCount.Add(ctx,
		1,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("Read")},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)})

	s.latencyHistogram.Record(ctx,
		latency,
		attribute.KeyValue{Key: "method", Value: attribute.StringValue("Read")},
		attribute.KeyValue{Key: "success", Value: attribute.BoolValue(err == nil)},
	)

	return data, err
}

// Close closes the connection pool.
func (s *InstrumentedSQLStorePGX) Close() {
	s.store.Close()
}
