.PHONY: default 
default: check

.PHONY: check 
check: check-quick test

.PHONY: check-quick
check-quick: adr format lint diagrams tidy

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
		golangci/golangci-lint:v2.2.2 golangci-lint run --color never

.PHONY: format
format:
	@docker run --rm -t -v $$(pwd):/app -w /app \
		-e GOCACHE=/.cache/go-build \
		-e GOMODCACHE=/.cache/mod \
		-e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
		-v ~/.cache/golagci-lint-docker:/.cache \
		golangci/golangci-lint:v2.2.2 golangci-lint fmt 

.PHONY: test
test:
	@go test -v ./...

.PHONY: tidy
tidy:
	@go mod tidy 

.PHONY: diagrams
diagrams:
	@docker run --rm -v ./docs:/docs plantuml/plantuml \
		-tsvg -o /docs/diagrams/dist /docs/diagrams

.PHONY: codegen
codegen: sqlc

.PHONY: sqlc
sqlc:
	@go tool sqlc generate

.PHONY: signoz-up
signoz-up:
	@docker compose \
		-f deployments/docker-compose/signoz/docker-compose.yaml \
		up -d

.PHONY: signoz-down
signoz-down:
	@docker compose \
		-f deployments/docker-compose/signoz/docker-compose.yaml \
		down

.PHONY: up
up: tidy format codegen
	@docker compose \
		-f deployments/docker-compose/docker-compose.yaml \
		up --build --force-recreate

.PHONY: down
down: 
	@docker compose \
		-f deployments/docker-compose/docker-compose.yaml \
		down 

# used for interactive development with tdd/bdd
.PHONY: delve-test
delve-test:
	@dlv test ./test
