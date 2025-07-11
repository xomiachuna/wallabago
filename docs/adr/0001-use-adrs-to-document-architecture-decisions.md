# 1. Use ADRs to document architecture decisions

Date: 2025-07-11

## Status

Accepted

## Context

We need to document various architectural descisions taken
during the design and development. There are multiple ways to track them - 
in an ad-hoc text file, colaborative document editors (e.g. Google Docs),
or next to the code using multiple markdown files as described in 
[Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)

## Decision

We will use ADRs to document architecture decisions.

## Consequences

The documentation of architecture decisions will now follow ADR format 
(`Status`, `Context`, `Decision`, `Consequences`) with some provenance 
(e.g. decisions being deprecated or superseded). This might require more effort
than keeping a single doc in a version control but results in a smaller, more
easily digestible documents in a format that is both easy to read (for humans 
and machines) and author.

In order to help with authoring and managing ADRs - a helper script is provided (see
[`tools/adr.sh`](../../tools/adr.sh) and [`adr-tools` repo](https://github.com/npryce/adr-tools))
