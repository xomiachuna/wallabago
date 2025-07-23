# 10. Use golang-migrate for database migrations

Date: 2025-07-23

## Status

Accepted

Influenced by [4. Use separate container for migrations](0004-use-separate-container-for-migrations.md)

Influenced by [9. Use sqlc for database query codegen](0009-use-sqlc-for-database-query-codegen.md)

Influenced by [8. Use PostgreSQL database](0008-use-postgresql-database.md)

## Context

We need to choose a tool that will run the database migrations. The tool needs
to be able to run in a separate container, support PostgreSQL and ideally interact
well with `sqlc`. We don't need to be able to automatically generate the migrations
based on the database, although automatic template generation is welcome.

Options include `golang-migrate` and `goose`.

## Decision

We will use `golang-migrate`.

## Consequences

We need to run `golang-migrate` container before the application starts/is updated.

If needed the migration process might be customized through a custom binary using
the library.

[Certain care](https://github.com/golang-migrate/migrate/blob/master/GETTING_STARTED.md#create-migrations) needs to be taken when authoring migrations in parallel on multiple branches
is order to ensure the proper merge later.
