package impl

import (
	"context"
	"fmt"

	"github.com/textileio/go-tableland/internal/tableland"
	"github.com/textileio/go-tableland/pkg/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
)

func (ep *EventProcessor) initMetrics(chainID tableland.ChainID) error {
	meter := global.MeterProvider().Meter("tableland")
	ep.mBaseLabels = append([]attribute.KeyValue{attribute.Int64("chain_id", int64(chainID))}, metrics.BaseAttrs...)

	// Async instruments.
	mExecutionRound, err := meter.AsyncInt64().Gauge("tableland.eventprocessor.execution.round")
	if err != nil {
		return fmt.Errorf("creating execution round gauge: %s", err)
	}

	mLastProcessedHeight, err := meter.AsyncInt64().Gauge("tableland.eventprocessor.last.processed.height")
	if err != nil {
		return fmt.Errorf("creating last processed height gauge: %s", err)
	}

	mHashCalculationElapsedTime, err := meter.AsyncInt64().Gauge("tableland.eventprocessor.hash.calculation.elapsed.time")
	if err != nil {
		return fmt.Errorf("creating hash calculation elapsed time gauge: %s", err)
	}

	mTreeLeavesCalculationElapsedTime, err := meter.
		AsyncInt64().
		Gauge("tableland.eventprocessor.tree.leaves.calculation.elapsed.time")
	if err != nil {
		return fmt.Errorf("creating hash calculation elapsed time gauge: %s", err)
	}

	err = meter.RegisterCallback([]instrument.Asynchronous{
		mExecutionRound, mLastProcessedHeight, mHashCalculationElapsedTime, mTreeLeavesCalculationElapsedTime,
	},
		func(ctx context.Context) {
			mExecutionRound.Observe(ctx, ep.mExecutionRound.Load(), ep.mBaseLabels...)
			mLastProcessedHeight.Observe(ctx, ep.mLastProcessedHeight.Load(), ep.mBaseLabels...)
			mHashCalculationElapsedTime.Observe(ctx, ep.mHashCalculationElapsedTime.Load(), ep.mBaseLabels...)
			mTreeLeavesCalculationElapsedTime.Observe(ctx, ep.mTreeLeavesCalculationElapsedTime.Load(), ep.mBaseLabels...)
		})
	if err != nil {
		return fmt.Errorf("registering async metric callback: %s", err)
	}

	// Sync instruments.
	ep.mEventExecutionCounter, err = meter.SyncInt64().Counter("tableland.eventprocessor.event.execution.count")
	if err != nil {
		return fmt.Errorf("creating event execution count instrument: %s", err)
	}
	ep.mTxnExecutionLatency, err = meter.SyncInt64().Histogram("tableland.eventprocessor.txn.execution.latency")
	if err != nil {
		return fmt.Errorf("creating txn execution latency instrument: %s", err)
	}
	ep.mBlockExecutionLatency, err = meter.SyncInt64().Histogram("tableland.eventprocessor.block.execution.latency")
	if err != nil {
		return fmt.Errorf("creating block execution latency instrument: %s", err)
	}

	return nil
}
