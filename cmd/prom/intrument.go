package main

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/aws/smithy-go/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// A promInstrument wraps a Prometheus metric and presents it
// as a [metrics.Instrument]. We defer construction of the metric
// until it is used, because we don't know if we have labels until then.
type promInstrument struct {
	init        sync.Once
	registry    prometheus.Registerer
	name        string
	description string
	metric      any
}

func collectInstrumentOptions(opts []metrics.InstrumentOption) *metrics.InstrumentOptions {
	o := &metrics.InstrumentOptions{}
	for _, fn := range opts {
		fn(o)
	}
	return o
}

func getLabels(opts []metrics.RecordMetricOption) map[any]any {
	o := &metrics.RecordMetricOptions{}
	for _, f := range opts {
		f(o)
	}
	return o.Properties.Values()
}

func getSortedKeys(lbls map[any]any) []string {
	var keys []string
	for k := range maps.Keys(lbls) {
		if s, ok := k.(string); ok {
			keys = append(keys, s)
		}
	}
	slices.Sort(keys)
	return keys
}

type instrumentType int

const (
	instrumentTypeCounter instrumentType = iota
	instrumentTypeGauge
	instrumentTypeHistogram
)

// instrumentName maps OTEL naming conventions to
// Prometheus naming conventions.
func instrumentName(name string, typ instrumentType, unitLabel string) string {
	name = fixName(name)

	// https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/translator/prometheus#metric-name

	unitStr := translateUnit(unitLabel, typ)

	if unitStr != "" {
		name = name + "_" + unitStr
	}

	if typ == instrumentTypeCounter {
		name = name + "_total"
	}

	return name
}

func translateUnit(s string, typ instrumentType) string {
	// special cases

	if strings.HasPrefix(s, "{") {
		return ""
	}

	ratioParts := strings.SplitN(s, "/", 2)
	if len(ratioParts) > 1 {
		return translateUnit(ratioParts[0], typ) + "_per_" + translateUnit(ratioParts[1], typ)
	}

	if s == "1" {
		if typ == instrumentTypeGauge {
			return "ratio"
		}
		return ""
	}

	// standard mapping
	punit, ok := unitTranslationMap[s]
	if ok {
		return punit
	}

	return fixName(s)
}

var unitTranslationMap = map[string]string{
	"d":   "days",
	"h":   "hours",
	"min": "minutes",
	"s":   "seconds",
	"ms":  "milliseconds",
	"us":  "microseconds",
	"ns":  "nanoseconds",

	"By":   "bytes",
	"KiBy": "kibibytes",
	"MiBy": "mebibytes",
	"GiBy": "gibibytes",
	"TiBy": "tibibytes",
	"KBy":  "kilobytes",
	"MBy":  "megabytes",
	"GBy":  "gigabytes",
	"TBy":  "terabytes",

	"m": "meters",
	"V": " volts",
	"A": "amperes",
	"J": "joules",
	"W": "watts",
	"g": "grams",

	"Cel": "celsius",
	"Hz":  "hertz",
	"%":   "percent",
}

var invalidCharRegexp = regexp.MustCompile("[^a-zA-Z0-9_:]")
var manyUnderneathies = regexp.MustCompile("__+")

func fixName(name string) string {
	// TODO - don't use regexes
	name = invalidCharRegexp.ReplaceAllString(name, "_")
	name = manyUnderneathies.ReplaceAllString(name, "_")
	return strings.Trim(name, "_")
}
