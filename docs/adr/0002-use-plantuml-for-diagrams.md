# 2. Use PlantUML for diagrams

Date: 2025-07-11

## Status

Accepted

Influenced by [1. Use ADRs to document architecture decisions](0001-use-adrs-to-document-architecture-decisions.md)

## Context

There is a need to keep track of certain diagrams (Entity Relationship,
Deployment, Sequence etc.) in order to keep the project structure understandable 
and well documented.

There are multiple ways to draw diagrams for software:
- ad-hoc image editing (e.g. on paper, in GIMP etc)
- using 3rd-party online tools (e.g. draw.io)
- use standalone software (e.g. editors for UML)
- use diagram from code generation tools (PlantUML, Mermaid etc.)

Since we use Markdown and ADRs - it would be beneficial to be able to render them
along with the ADRs.

It would also be good to be able to track changes between diagram versions, preferably
in a textual and visual format.

Primary contenders thus are Mermaid and PlantUML. Their feature-set is different
with PlantUML being a more mature option but Mermaid being the only one supported
by GitHub at the moment.

## Decision

We will use PlantUML.

## Consequences

Architecture description will now rely on diagrams written in PlantUML and converted to images.
The diagrams need to be updated based on the contents of the source `.puml` files.

For development purposes a plantuml rendering server is necessary, with available options being
an [online editor](https://editor.plantuml.com/) (which requires manual saving of the files to the
source tree), a [vscode extension](https://marketplace.visualstudio.com/items?itemName=jebbs.plantuml) for
interactive editing and preview and a standalone local plantuml deployment.

In order to guarantee the proper sync between `.puml` source files and diagrams a CI tool integration
is needed, which might make the integration process more brittle and complicated.
