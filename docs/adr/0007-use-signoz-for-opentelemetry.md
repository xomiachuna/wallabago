# 7. Use SigNoz for OpenTelemetry

Date: 2025-07-16

## Status

Proposed

Influenced by [6. Use OpenTelemetry instrumentation](0006-use-opentelemetry-instrumentation.md)

## Context

We need to choose a Collector implementation for OpenTelemetry.
Ideally the features would be:
- support for metrics, traces and logs and correlations between them
- ability to run queries
- web ui
- open source

## Decision

We will use SigNoz for OpenTelemetry.

## Consequences

Signoz needs to be deployed along with its dependencies, potentially complicating
the setup.

No vendor lock-in allows to replace it with a different tool if needed.

Logs and traces now can be collected in a single place and correlated and queried.

Metrics can be exported as well.
