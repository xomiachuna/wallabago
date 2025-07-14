# 4. Use separate container for migrations

Date: 2025-07-14

## Status

Proposed

## Context

There are two main ways to run migrations:

- from within the application container at startup
- from a separate container (possibly independently from main app)

Having a single container without replicas is easy, but once you move to
multiple replicas the migration process becomes more brittle due to potential
concurrent migrations applying the same changes, requiring the migrations to be
written as idempotent.

Having a separate container for migrations allows to have a separate lifecycle
for migrations but requires them to be done in a backward-compatible manner in
order to not interrupt the running instances with breaking schema changes.

> A.K.: currently there is also an educational opportunity - learning to work in
> a multi-replica environment is more challenging but will happen more often in
> real world scenarios, so I'm considering it from this standpoint

## Decision

We will use a standalone migration container.

## Consequences

Migrations need to be implemented in backwards-compatible manner. Potentially
breaking changes need to be planned ahead and guarder off with a feature flag.
The application startup process is now slower if it includes schema changes. An
additional dependency on migrate/migrate increases the surface area of the
stack, making the setup potentially more brittle.
