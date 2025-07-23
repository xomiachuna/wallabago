# 8. Use PostgreSQL database

Date: 2025-07-18

## Status

Accepted

Influences [9. Use sqlc for database query codegen](0009-use-sqlc-for-database-query-codegen.md)

Influences [10. Use golang-migrate for database migrations](0010-use-golang-migrate-for-database-migrations.md)

## Context

There are many databases that might be suitable for the application.
Given the mostly CRUD nature of the app and anticipated relative stability of the
schema, SQL databases are a natural choice.

Among the anticipated features the following are taken into consideration:
- maturity of the ecosystem and tooling
- support for extensions (e.g. vector search)
- simple operation
- no vendor lock-in
- good performance

The options basically boil down to well-established RDBMS like PostgreSQL and SQLite

## Decision

We will use PostgreSQL as the primary database.

## Consequences

We need to choose a `database/sql` driver that works with postgres and can be
instrumented using Otel.

We need to choose the method of accessing the database - hand-rolled query processing,
codegen or an ORM.

Database Migration tooling needs to support PostgreSQL.

Operation of the database requires additional care taken to monitor its availability
as well as backups.
