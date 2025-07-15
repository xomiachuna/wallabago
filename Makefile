.PHONY: default 
default: check

.PHONY: check 
check: check-quick

.PHONY: check-quick
check-quick: adr format lint diagrams

.PHONY: adr
adr: generate-adr-toc generate-adr-graph 

.PHONY: generate-adr-graph
generate-adr-graph:
	@./tools/adr.sh generate graph \
		| docker run --rm -i nshine/dot dot -Tpng -Gdpi=300 \
		> docs/adr/adr.png

.PHONY: generate-adr-toc
generate-adr-toc:
	@./tools/adr.sh \
		generate toc \
		-i /doc/adr/intro.template.md \
		> docs/adr/README.md

.PHONY: lint
lint: format
	@docker run --rm -t -v $$(pwd):/app -w /app \
		-e GOCACHE=/.cache/go-build \
		-e GOMODCACHE=/.cache/mod \
		-e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
		-v ~/.cache/golagci-lint-docker:/.cache \
		golangci/golangci-lint:v2.2.2 golangci-lint run

.PHONY: format
format:
	@docker run --rm -t -v $$(pwd):/app -w /app \
		-e GOCACHE=/.cache/go-build \
		-e GOMODCACHE=/.cache/mod \
		-e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
		-v ~/.cache/golagci-lint-docker:/.cache \
		golangci/golangci-lint:v2.2.2 golangci-lint fmt

.PHONY: diagrams
diagrams:
	@docker run --rm -v ./docs:/docs plantuml/plantuml \
		-tsvg -o /docs/diagrams/dist /docs/diagrams

.PHONY: up
up:
	@docker compose \
		-f deployments/docker-compose/docker-compose.yaml \
		up --build --force-recreate

.PHONY: down
down:
	@docker compose \
		-f deployments/docker-compose/docker-compose.yaml \
		down 
