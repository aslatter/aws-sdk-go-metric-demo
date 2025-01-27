package main

import (
	"context"

	"github.com/aws/smithy-go/metrics"
)

type noopInstrumentFloat struct{}

// Sample implements metrics.Float64Gauge.
func (n *noopInstrumentFloat) Sample(context.Context, float64, ...metrics.RecordMetricOption) {
	// noop
}

// Add implements metrics.Float64Counter.
func (n *noopInstrumentFloat) Add(context.Context, float64, ...metrics.RecordMetricOption) {
	// noop
}

// Stop implements metrics.AsyncInstrument.
func (n *noopInstrumentFloat) Stop() {
	// noop
}

type noopInstrumentInt struct{}

// Record implements metrics.Int64Histogram.
func (n *noopInstrumentInt) Record(context.Context, int64, ...metrics.RecordMetricOption) {
	// noop
}

// Sample implements metrics.Int64Gauge.
func (n *noopInstrumentInt) Sample(context.Context, int64, ...metrics.RecordMetricOption) {
	// noop
}
