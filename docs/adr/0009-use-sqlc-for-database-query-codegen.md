# 9. Use sqlc for database query codegen

Date: 2025-07-18

## Status

Accepted

Influenced by [8. Use PostgreSQL database](0008-use-postgresql-database.md)

Influences [10. Use golang-migrate for database migrations](0010-use-golang-migrate-for-database-migrations.md)

## Context

There are multiple approaches to writing applications that use a SQL database:
- hand-rolled quries: prone to repetition but provides the most amount of control, low abstraction
- orm: provides less control over the queries, risk of hidden N+1 problems, high abstraction
- codegen: more control than orm, less control than hand-rolled queries, low-cost medium abstraction
(can be easily introspected by reading the generated code)

Tooling options:
- hand-rolled: `pgx`, `database/sql`
- orm: `gorm`
- codegen: `sqlc`

## Decision

We will use `sqlc` for codegen

## Consequences

Codegen becomes part of the build process.

Queries can be written as optimally as necessary - along with prepared statements.

`sqlc` needs to be confgured with the db migration solution as to not interfere negatively
with each other.
