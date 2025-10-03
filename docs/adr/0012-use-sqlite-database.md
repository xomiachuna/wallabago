# 12. Use SQLite Database

Date: 2025-10-03

## Status

Accepted

Supercedes [8. Use PostgreSQL database](0008-use-postgresql-database.md)

Influenced by [9. Use sqlc for database query codegen](0009-use-sqlc-for-database-query-codegen.md)

## Context

Currently the deployment and operation of the application has a dependency on external infra in the form
of a postgres database. While separation is nice, the operational overhead makes us depend on
docker for the purpose of simply running the app. This complicates the testing setup (we need to use testcontainers/compose).
The scale that can be handled by postgres is not realistically to be achieved by any installation
of this app (which will at most have a few users). We are not entirely bound by database interaction layer 
as we dont use any PG-specific features (apart from schemas which are used only to provide 'namespaces' for db tables
which we intend to keep loosely coupled - namely idp stuff). Since we use SQLC the cost of migrating to other
SQL-compatible database is minimized.

## Decision

We will use SQLite database.

## Consequences

The database layer needs to be reimplemented using SQLite flavor of SQL. We need to enable
sane settings for SQLite (type safety, wal etc).

Migrations need to be integrated in the main binary (as opposed to running migrate separately as it
is done now).

The infra setup can be simplified - we will be able to ditch docker-compose for the main use case 
(app without otel). This would allow us to drop testcontainers (at least until we dont use otel in
tests) and have a single deliverable binary. For portability sake we will keep the Docker image -
it will be used as a reproducible build environment rather than the main target platform (as well as
potentially being used to test otel in the future).

Since we dont have any live installations, this should be a painless migration.

We need to consider replacing Postgres schemas either with a naming convention for the tables.
