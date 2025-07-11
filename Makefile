.PHONY: default 
default: 
	@# run check rule with max parallelism
	@$(MAKE) -j --no-print-directory check

.PHONY: check 
check: adr lint

.PHONY: adr
adr: generate-adr-graph generate-adr-toc

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
	@go fmt ./...
