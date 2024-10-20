package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/metrics"

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
	meterProvider := newMeterProvider(&meterProviderOptions{
		registry:  promRegistry,
		namespace: "aws",
	})

	// for demo purposes, scrape all prom metrics and dump to stdout
	defer scrapePromMetrics(promRegistry)

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("loading aws config: %s", err)
	}
	cfg.Region = "us-east-1"

	err = callS3(ctx, meterProvider, cfg)
	if err != nil {
		return err
	}

	err = callDynamoDB(ctx, meterProvider, cfg)
	if err != nil {
		return err
	}

	return nil
}

func callS3(ctx context.Context, meterProvider metrics.MeterProvider, cfg aws.Config) error {
	s3c := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.MeterProvider = meterProvider
	})

	// make a few API calls

	_, err := s3c.ListBuckets(ctx, &s3.ListBucketsInput{})
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

func callDynamoDB(ctx context.Context, meterProvider metrics.MeterProvider, cfg aws.Config) error {
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.MeterProvider = meterProvider
	})

	_, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return err
	}

	_, err = client.ListGlobalTables(ctx, &dynamodb.ListGlobalTablesInput{})

	return err
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
