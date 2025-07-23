# wallabago - an implementation of Wallabag in golang

## Development

### Setup

This project requires `go>=1.24`, `docker` and `make`.

This project uses [`pre-commit`](https://pre-commit.com) in
order to run pre-commit hooks defined in [`.pre-commit-config.yaml`](./.pre-commit-config.yaml).
Install `pre-commit` (e.g. via `pipx install pre-commit`) and run:
```
pre-commit install
```

### Running in `docker-compose`
> [!NOTE]
>
> See [`deployments/docker-compose/`](./deployments/docker-compose/)

```
make up # make down for stopping
```
## Documentation
See [`docs/`](./docs/).

