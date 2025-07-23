# 6. Use OpenTelemetry instrumentation

Date: 2025-07-16

## Status

Proposed

Influences [7. Use SigNoz for OpenTelemetry](0007-use-signoz-for-opentelemetry.md)

## Context

There are multiple ways to instrument application code. The tooling for that ideally
should have as little vendor lock-in as possible with a large ecosystem of tooling
around it.

## Decision

We are going to use OpenTelemetry for instrumentation

## Consequences

Application code now may be instrumented using OpenTelemetry spans. Distributed
tracing is now available due to support of OpenTelemetry across many tools like
databases and libraries.

The HTTP server initial setup becomes more complicated due to the need to configure
OpenTelemetry and initialize the root spans as necessary.

Deployment of OpenTelemetry collector is necessary (alternatively can be replaced
using CLI/File exporter in order to simplify setup)
