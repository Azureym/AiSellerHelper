# Self documented Makefile
# http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show list of make targets and their description
	@grep -E '^[/%.a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL:= help

.PHONY: setup
setup: ## Run setup scripts to prepare development environment
	@scripts/setup.sh

.PHONY: clean
clean: ## Clean project dir, remove build artifacts and logs
	@scripts/clean.sh

.PHONY: test
test: ## Generate mock and run all test. To run specified tests, use `./scripts/test.sh <pattern>`)
	@scripts/test.sh $*

.PHONY: lint
lint: ## Run linter with --fast option, for most common issues
	@scripts/lint.sh

.PHONY: lint/all
lint/all: ## Run all the linter, slow, but stricter
	@scripts/lint.sh all

.PHONY: gen
gen: ## Show gen.sh help
	@scripts/gen.sh

gen/%: ## Generate artifacts that defined by '%', e.g: 'make gen.proto` will trigger ./scripts/gen.sh proto
	@scripts/gen.sh $*

.PHONY: build
build: ## Show build.sh help for building binary package under cmd
	@scripts/build.sh

build/%: ## Build artifact defined by '%', e.g: 'make build.server` will trigger ./scripts/build.sh server
	@scripts/build.sh $*

.PHONY: docs
docs: ## Build document for the proto files
	@scripts/docs.sh

all: clean setup gen/all build/all  ## Clean, setup, generate and then build all the binaries.

release: ## Build minimal size binaries for building docker image or deploying
	@# -s and -w linker flags will remove DWARF section in the binary, with is
	@#  only used for debugger. For binaries that run on production, we don't
	@#  really have the luxuxy to debug it, thus, it is safe to remove them
	@GO_LDFLAGS="-s -w" make gen/go gen/code build/all
