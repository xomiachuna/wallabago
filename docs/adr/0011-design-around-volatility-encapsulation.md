# 11. Design around volatility encapsulation

Date: 2025-07-23

## Status

Accepted

## Context

There are multiple ways to approach software design, particularly the internal
structure of it. Some approaches are concerned with the tiering of data-logic-presentation,
some are built around the domain concepts, some prioritize the technical tooling like message
passing/queues as the central substrate on which everything else works.

In order to have a somewhat consistent development experience we need to choose an approach that
roughly satisfies the following criteria:
- reasonable level of abstraction
- prescriptive guidance around the structure
- little to no restrictions on the tooling used
- resilience to change

> [!NOTE]
> 
> This is mostly based on personal preference: change is inevitable, minimizing blast
> radius of it simplifies the development process due to pragmatic design to encapsulate it

## Decision

We are going to use a design process centered around encapsulating volatility
as described in [__Righting Software__ by Juval Lowy](https://share.google/ESqILqfLfVAgVo9RG).

## Consequences

The application is structured around 4 primary Tiers: Data Storage (particular storage interfaces), 
Data Access (an abstraction around how specific data is stored - potentially hiding multiple
storage type interactions), Engine (an interface for a complicated but atomic in
a business-logic sense set of operations, potentially utilizing one or more Data Acccess
tier objects) and a Manager (the orchestration layer that coordinates the implementation
of particular use-cases as an interaction between Data Access and Engine tiers).

Finally the outer layer of the application code is the presentation/transport - 
e.g. HTTP handlers, gRPC interface etc, which expose the Managers to the outside world.

There are multiple restrictions to this approach (no horizontal or upward calls,
managers being lightweight etc.) that are described in-depth in __Righting Software__.

The resulting designs need to be visualized in a form of component diagram.
