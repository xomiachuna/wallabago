# wallabago - an implementation of Wallabag in golang

## Development
### Setup
This project requires `go>=1.24`.

This project uses [`lefthook`](https://github.com/evilmartians/lefthook) in order
to run pre-commit hooks defined in [`lefthook.yml`](./lefthook.yml). It can be ran via 
`go tool lefthook`.

Run the following in order to make sure that your local env has `lefthook` installed 
and configured:
```
go tool lefthook install && go tool lefthook run pre-commit
```

## Documentation
### ADRs
This project uses [ADR](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) format
to keep track of architecture decisions (see [`docs/adr/`](./docs/adr/) folder).

A [helper script](./tools/adr.sh) is provided ro run [`adr-tools`](https://github.com/npryce/adr-tools).
Run `./tools/adr.sh` from the root of the project in order to interact with the ADRs
