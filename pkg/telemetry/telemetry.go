package telemetry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	logger "github.com/rs/zerolog/log"
)

var (
	metricStore MetricStore
	log         zerolog.Logger

	mu   = &sync.Mutex{}
	once sync.Once
)

func init() {
	log = logger.With().
		Str("component", "telemetry").
		Logger()
}

// MetricStore specifies the methods for persisting a metric.
type MetricStore interface {
	StoreMetric(context.Context, Metric) error
	Close() error
}

// SetMetricStore sets the store implementation.
// Only the first call will have an effect. If Collect is called without setting a MetricStore, it will be a noop.
func SetMetricStore(s MetricStore) {
	once.Do(func() {
		metricStore = s
	})
}

// Collect collects the metric by persisting locally for later publication.
// If Collect is called before setting the metric store, it will simply log the metric without persisting it.
func Collect(ctx context.Context, metric interface{}) error {
	mu.Lock()
	defer mu.Unlock()
	if metricStore == nil {
		log.Warn().Msg("no metric store was set")
		return nil
	}

	switch v := metric.(type) {
	case StateHashMetric:
		if err := metricStore.StoreMetric(ctx, Metric{
			Version:   1,
			Timestamp: time.Now().UTC(),
			Type:      StateHashType,
			Payload: StateHashMetric{
				Version:     v.Version,
				ChainID:     v.ChainID,
				BlockNumber: v.BlockNumber,
				Hash:        v.Hash,
			},
		}); err != nil {
			return errors.Errorf("store state hash metric: %s", err)
		}
		return nil
	case GitSummaryMetric:
		if err := metricStore.StoreMetric(ctx, Metric{
			Version:   1,
			Timestamp: time.Now().UTC(),
			Type:      GitSummaryType,
			Payload: GitSummaryMetric{
				Version:       v.Version,
				GitCommit:     v.GitCommit,
				GitBranch:     v.GitBranch,
				GitState:      v.GitState,
				GitSummary:    v.GitSummary,
				BuildDate:     v.BuildDate,
				BinaryVersion: v.BinaryVersion,
			},
		}); err != nil {
			return errors.Errorf("store git summary metric: %s", err)
		}
		return nil
	case ChainStacksMetric:
		if err := metricStore.StoreMetric(ctx, Metric{
			Version:   1,
			Timestamp: time.Now().UTC(),
			Type:      ChainStacksSummaryType,
			Payload: ChainStacksMetric{
				Version:                   v.Version,
				LastProcessedBlockNumbers: v.LastProcessedBlockNumbers,
			},
		}); err != nil {
			return errors.Errorf("store chains stacks summary metric: %s", err)
		}
		return nil
	case ReadQueryMetric:
		if err := metricStore.StoreMetric(ctx, Metric{
			Version:   1,
			Timestamp: time.Now().UTC(),
			Type:      ReadQueryType,
			Payload: ReadQueryMetric{
				Version:       v.Version,
				IPAddress:     v.IPAddress,
				SQLStatement:  v.SQLStatement,
				FormatOptions: v.FormatOptions,
				TookMilli:     v.TookMilli,
			},
		}); err != nil {
			return errors.Errorf("read query metric: %s", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown metric type %T", v)
	}
}
