package main

import (
	"sync"

	"github.com/aws/smithy-go/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var _ metrics.MeterProvider = (*promMeterProvider)(nil)

// A promMeterProvider is an adapter which maps
// Prometheus metrics to the smithy-go metric interfaces.
//
// This is tricksy because:
//
//   - The AWS SDKs rely on the instrument-caching behavior
//     of the OTEL meter provider
//   - Metric-labels are not provided until a metric is
//     observed.
//
// So we do caching and delayed instantiation.
type promMeterProvider struct {
	prefix   string
	registry prometheus.Registerer
	// The OTEL meter-provider caches instruments, and the AWS SDK
	// assumes this behavior. The prometheus client does not do this
	// natively.
	metricCache cache
}

type meterProviderOptions struct {
	namespace string
	registry  prometheus.Registerer
}

func newMeterProvider(opts *meterProviderOptions) *promMeterProvider {

	r := opts.registry
	if r == nil {
		r = prometheus.DefaultRegisterer
	}

	var prefix string
	if opts.namespace != "" {
		prefix = opts.namespace + "_"
	}

	return &promMeterProvider{
		registry: r,
		prefix:   prefix,
	}
}

// Meter implements metrics.MeterProvider.
func (a *promMeterProvider) Meter(scope string, opts ...metrics.MeterOption) metrics.Meter {
	// TODO - optionally add scope as a label? It would need to be included in the cache-key
	return &promMeter{
		parent: a,
	}
}

var _ metrics.Meter = (*promMeter)(nil)

type promMeter struct {
	parent *promMeterProvider
}

// Float64AsyncCounter implements metrics.Meter.
func (p *promMeter) Float64AsyncCounter(name string, callback metrics.Float64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	panic("unimplemented")
}

// Float64AsyncGauge implements metrics.Meter.
func (p *promMeter) Float64AsyncGauge(name string, callback metrics.Float64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	panic("unimplemented")
}

// Float64AsyncUpDownCounter implements metrics.Meter.
func (p *promMeter) Float64AsyncUpDownCounter(name string, callback metrics.Float64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	panic("unimplemented")
}

// Float64Counter implements metrics.Meter.
func (p *promMeter) Float64Counter(name string, opts ...metrics.InstrumentOption) (metrics.Float64Counter, error) {
	panic("unimplemented")
}

// Float64Gauge implements metrics.Meter.
func (p *promMeter) Float64Gauge(name string, opts ...metrics.InstrumentOption) (metrics.Float64Gauge, error) {
	panic("unimplemented")
}

// Float64Histogram implements metrics.Meter.
func (p *promMeter) Float64Histogram(name string, opts ...metrics.InstrumentOption) (metrics.Float64Histogram, error) {
	m := p.getInstrument(name, opts)
	return (*histogramInstrument[float64])(m), nil
}

// Float64UpDownCounter implements metrics.Meter.
func (p *promMeter) Float64UpDownCounter(name string, opts ...metrics.InstrumentOption) (metrics.Float64UpDownCounter, error) {
	panic("unimplemented")
}

// Int64AsyncCounter implements metrics.Meter.
func (p *promMeter) Int64AsyncCounter(name string, callback metrics.Int64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	panic("unimplemented")
}

// Int64AsyncGauge implements metrics.Meter.
func (p *promMeter) Int64AsyncGauge(name string, callback metrics.Int64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	panic("unimplemented")
}

// Int64AsyncUpDownCounter implements metrics.Meter.
func (p *promMeter) Int64AsyncUpDownCounter(name string, callback metrics.Int64Callback, opts ...metrics.InstrumentOption) (metrics.AsyncInstrument, error) {
	panic("unimplemented")
}

// Int64Counter implements metrics.Meter.
func (p *promMeter) Int64Counter(name string, opts ...metrics.InstrumentOption) (metrics.Int64Counter, error) {
	m := p.getInstrument(name, opts)
	return (*counterInstrument[int64])(m), nil
}

// Int64Gauge implements metrics.Meter.
func (p *promMeter) Int64Gauge(name string, opts ...metrics.InstrumentOption) (metrics.Int64Gauge, error) {
	panic("unimplemented")
}

// Int64Histogram implements metrics.Meter.
func (p *promMeter) Int64Histogram(name string, opts ...metrics.InstrumentOption) (metrics.Int64Histogram, error) {
	panic("unimplemented")
}

// Int64UpDownCounter implements metrics.Meter.
func (p *promMeter) Int64UpDownCounter(name string, opts ...metrics.InstrumentOption) (metrics.Int64UpDownCounter, error) {
	m := p.getInstrument(name, opts)
	return (*gaugeInstrument[int64])(m), nil
}

// todo - noop instrument for unimplemented stuff?

// getInstrument returns a previously cached instrument or
// instantiates and caches a new one.
func (p *promMeter) getInstrument(name string, opts []metrics.InstrumentOption) *promInstrument {
	o := collectInstrumentOptions(opts)
	name = p.parent.prefix + instrumentName(name, instrumentTypeCounter, o.UnitLabel)

	m := p.parent.metricCache.lookupOrInsert(name, func() *promInstrument {
		return &promInstrument{
			name:        name,
			description: o.Description,
			registry:    p.parent.registry,
		}
	})

	return m
}

type cache struct {
	l sync.Mutex
	// TODO - sync.Map would likely have less contention given how this map is used
	m map[string]*promInstrument
}

func (c *cache) lookupOrInsert(k string, mk func() *promInstrument) *promInstrument {
	c.l.Lock()
	defer c.l.Unlock()

	metric, ok := c.m[k]
	if ok {
		return metric
	}

	if c.m == nil {
		c.m = map[string]*promInstrument{}
	}

	metric = mk()
	c.m[k] = metric

	return metric
}
