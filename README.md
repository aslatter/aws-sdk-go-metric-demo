
This project is a small programming exercise which demonstrates
using the built-in metrics in aws-sdk-go-v2 but exposing them
as a Prometheus scrape-target.

The AWS SDK clients optionally take-in a "meter provider", which
is an implementation of an AWS-supplied interface, and the Go
"smithy" libraries (provided by Amazon) have an adapter to allow
using the Go OTEL metrics libraries with the AWS SDK.

The package `./cmd/otel` wires together:

+ The AWS SDK
+ The AWS adapter for the OTEL metrics SDK
+ The OTEL Prometheus exporter

This is how I'd recommend integrating the Prometheus client
library with the AWS SDK.

The package `./cmd/prom` takes a different approach - it provides
a "Prometheus native" implementation of the AWS SDK meter-provider.

It's a pain to do and probably not worth it.
