package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	s3c := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.MeterProvider = smithyotelmetrics.Adapt(meterProvider)
	})

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
