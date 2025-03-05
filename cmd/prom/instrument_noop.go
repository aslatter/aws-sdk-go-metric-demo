package main

import (
	"context"

	"github.com/aws/smithy-go/metrics"
)

type noopInstrument[T float64 | int64] struct{}

// Add implements metrics.{FloatInt}64UpDownCounter.
func (n *noopInstrument[T]) Add(context.Context, T, ...metrics.RecordMetricOption) {
	// noop
}

// Stop implements metrics.AsyncInstrument.
func (n *noopInstrument[T]) Stop() {
	// noop
}

// Sample implements metrics.{Float|Int}64Gauge.
func (n *noopInstrument[T]) Sample(context.Context, T, ...metrics.RecordMetricOption) {
	// noop
}

// Record implements metrics.{Float|Int}64Histogram.
func (n *noopInstrument[T]) Record(context.Context, T, ...metrics.RecordMetricOption) {
	// noop
}
