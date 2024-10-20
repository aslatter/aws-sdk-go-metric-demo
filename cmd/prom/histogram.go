package main

import (
	"context"

	"github.com/aws/smithy-go/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type histogramInstrument[T float64 | int64] promInstrument

// Record implements metrics.Float64Histogram.
func (f *histogramInstrument[T]) Record(ctx context.Context, v T, opts ...metrics.RecordMetricOption) {
	lbls := getLabels(opts)
	// TODO - cache sorted keys after first invocation?
	keys := getSortedKeys(lbls)

	f.init.Do(func() {

		// do we have labels?
		if len(keys) > 0 {
			var fixedKeys []string
			for _, k := range keys {
				fixedKeys = append(fixedKeys, fixName(k))
			}

			m := prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    f.name,
					Help:    f.description,
					Buckets: prometheus.DefBuckets,
				},
				fixedKeys,
			)
			f.registry.MustRegister(m)
			f.metric = m

			return
		}

		// otherwise a plain histogram
		m := prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    f.name,
			Help:    f.description,
			Buckets: prometheus.DefBuckets,
		})
		f.registry.MustRegister(m)
		f.metric = m
	})

	if len(keys) > 0 {
		h, ok := f.metric.(*prometheus.HistogramVec)
		if !ok {
			// :(
			return
		}
		var vals []string
		for _, k := range keys {
			l, _ := lbls[k].(string)
			vals = append(vals, l)
		}
		h.WithLabelValues(vals...).Observe(float64(v))
		return
	}

	h, ok := f.metric.(prometheus.Histogram)
	if !ok {
		return
	}
	h.Observe(float64(v))
}
