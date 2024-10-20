package main

import (
	"context"

	"github.com/aws/smithy-go/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type gaugeInstrument[T float64 | int64] promInstrument

// Add implements metrics.Int64Counter.
func (i *gaugeInstrument[T]) Add(ctx context.Context, v T, opts ...metrics.RecordMetricOption) {
	lbls := getLabels(opts)
	// TODO - cache sorted keys after first invocation?
	keys := getSortedKeys(lbls)

	i.init.Do(func() {

		// do we have labels?
		if len(keys) > 0 {
			var fixedKeys []string
			for _, k := range keys {
				fixedKeys = append(fixedKeys, fixName(k))
			}

			m := prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: i.name,
					Help: i.description,
				},
				fixedKeys,
			)
			i.registry.MustRegister(m)
			i.metric = m

			return
		}

		// otherwise a plain histogram
		m := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: i.name,
			Help: i.description,
		})
		i.registry.MustRegister(m)
		i.metric = m
	})

	if len(keys) > 0 {
		h, ok := i.metric.(*prometheus.GaugeVec)
		if !ok {
			// :(
			return
		}
		var vals []string
		for _, k := range keys {
			l, _ := lbls[k].(string)
			vals = append(vals, l)
		}
		h.WithLabelValues(vals...).Add(float64(v))
		return
	}

	h, ok := i.metric.(prometheus.Gauge)
	if !ok {
		return
	}
	h.Add(float64(v))
}