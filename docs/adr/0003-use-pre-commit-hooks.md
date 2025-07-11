# 3. Use lefthook for pre-commit hooks

Date: 2025-07-11

## Status

Proposed

Influenced by [2. Use PlantUML for diagrams](0002-use-plantuml-for-diagrams.md)

## Context

There are multiple invariants that we'd like to preserve as much as possible
when committing changes - formatting, linting, diagrams being syncronized with the
source etc. There are tools that allow to change the files in-place in order
to preserve those invariants. Some tools (like plantuml) might require/benefit from
being able to be ran in a docker container with the source tree being mounted as
a volume - this is simpler and more portable than requiring the tools to be installed
on the dev machine.

Git provides a way to add hooks to commits, which will be ran before the commit is
finalized.

There are tools that allow to declare the hooks that need to be ran: `pre-commit`, 
`lefthook`, `huskey` etc.
They have similar features, different configurations and ecosystems surrounding them.

We need to balance the features with simplicity and reliability.

## Decision

We will use [lefthook](https://github.com/evilmartians/lefthook) bundled as a 
[tool dependency](https://tip.golang.org/doc/modules/managing-dependencies#tools)

## Consequences

> TODO
