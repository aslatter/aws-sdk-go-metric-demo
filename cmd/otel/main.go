package main

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/metrics"
	"github.com/aws/smithy-go/metrics/smithyotelmetrics"

	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	// set up our metric-exporter
	promRegistry := prometheus.NewRegistry()
	meterProvider := setupOTELExporter(promRegistry)

	// for demo purposes, scrape all prom metrics and dump to stdout
	defer scrapePromMetrics(promRegistry)

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("loading aws config: %s", err)
	}

	addMeterProvider(&cfg, smithyotelmetrics.Adapt(meterProvider))

	s3c := s3.NewFromConfig(cfg)

	// make a few API calls

	_, err = s3c.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("list buckets: %s", err)
	}

	_, err = s3c.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("list buckets: %s", err)
	}

	_, err = s3c.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("list buckets: %s", err)
	}

	return nil
}

// setupOTELExporter creates an OTEL meter-provider whose back-end is the
// provided Prometheus registry.
func setupOTELExporter(promRegistry *prometheus.Registry) *sdkmetric.MeterProvider {
	// create an otel metric-exporter associated with the
	// default prometheus registry
	metricExporter, err := otelprom.New(
		otelprom.WithNamespace("aws"),
		otelprom.WithoutScopeInfo(),
		otelprom.WithoutTargetInfo(),
		otelprom.WithRegisterer(promRegistry),

		// OTEL default buckets assume you're using milliseconds. Substitute defaults
		// appropriate for units of seconds.
		//
		// https://github.com/open-telemetry/opentelemetry-go/issues/5821
		otelprom.WithAggregationSelector(func(ik sdkmetric.InstrumentKind) sdkmetric.Aggregation {
			switch ik {
			case sdkmetric.InstrumentKindHistogram:
				return sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: prometheus.DefBuckets,
					NoMinMax:   false,
				}
			default:
				return sdkmetric.DefaultAggregationSelector(ik)
			}
		}),
	)
	if err != nil {
		panic(err)
	}

	// create a meter-provider associated with the exporter
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(metricExporter),
		// opt-out of metrics which appear to be purely client-side computation
		sdkmetric.WithView(func(i sdkmetric.Instrument) (sdkmetric.Stream, bool) {
			switch i.Name {
			case "client.call.serialization_duration", "client.call.deserialization_duration", "client.call.resolve_endpoint_duration", "client.call.auth.signing_duration":
				return sdkmetric.Stream{Aggregation: sdkmetric.AggregationDrop{}}, true
			}
			return sdkmetric.Stream{}, false
		}),
	)

	return meterProvider
}

// for demo purposes, dump all prom metrics to stdout
func scrapePromMetrics(promRegistry *prometheus.Registry) {
	metricFamilies, err := promRegistry.Gather()
	if err != nil {
		panic(err)
	}

	encoder := expfmt.NewEncoder(os.Stdout, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, mf := range metricFamilies {
		if err := encoder.Encode(mf); err != nil {
			panic(err)
		}
	}
}

// well this is gross.
// https://github.com/aws/aws-sdk-go-v2/issues/2927
func addMeterProvider(cfg *aws.Config, meterProvider metrics.MeterProvider) {
	mpv := reflect.ValueOf(meterProvider)
	noopType := reflect.TypeFor[metrics.NopMeterProvider]()
	cfg.ServiceOptions = append(cfg.ServiceOptions, func(s string, a any) {
		v := reflect.ValueOf(a)
		// we expect 'a' to be a pointer to a struct with a
		// 'MeterProvider' field of type 'metrics.MeterProvider'
		if v.Kind() != reflect.Pointer {
			return
		}
		ve := v.Elem()

		if ve.Kind() != reflect.Struct {
			return
		}

		f := ve.FieldByName("MeterProvider")
		if !f.IsValid() {
			return
		}
		if !f.CanSet() {
			return
		}

		// check that the field is empty. the SDK will initialize
		// empty meter-providers with a "no op" meter provider, so we
		// check for that, too.
		if !f.IsZero() && f.Kind() == reflect.Interface && f.Elem().Type() != noopType {
			return
		}

		if !mpv.Type().AssignableTo(f.Type()) {
			return
		}

		f.Set(mpv)
	})
}
