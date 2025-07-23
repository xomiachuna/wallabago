# 5. Use pre-commit for pre-commit hooks

Date: 2025-07-15

## Status

Accepted

Supercedes [3. Use lefthook for pre-commit hooks](0003-use-pre-commit-hooks.md)

## Context

`lefthook` has an [issue with not including or failing when hooks update additional
files not originally staged](https://github.com/evilmartians/lefthook/discussions/580),
which leads to multiple commits having to be created in case of user error (not cancelling
the commit and adding the newly staged changes).

Additionally this will be tricky to use with CI/CD as it might skip some changes.

## Decision

We will use [`pre-commit`](https://pre-commit.com/)

## Consequences

In order to use the pre-commit feature developers will now need to have a distribution
of python and pre-commit installed. `lefthook` go tool will be removed.

This should prevent changes from passing pre-commit checks locally until the checks
result in no additional changes.
