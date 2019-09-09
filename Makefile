PKG_SPEC = ./registry/...
MOD = -mod=readonly
GOTEST = go test $(MOD)
GOTEST_COVERAGE_OPT = -coverprofile=coverage.txt -covermode=atomic
TOOLS_DIR = $(shell git rev-parse --show-toplevel)/.tools
GOBUILD = go build $(MOD)

ENV ?= development
LINT_RUN_OPTS ?= --fix
override GOTEST_OPT += -timeout 30s

.DEFAULT_GOAL := precommit

$(TOOLS_DIR)/golangci-lint: go.mod go.sum tools.go
	$(GOBUILD) -o $(TOOLS_DIR)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: vars
vars:
	@echo "PKG_SPEC=$(PKG_SPEC)"
	@echo "MOD=$(MOD)"
	@echo "GOTEST=$(GOTEST)"
	@echo "COVERAGE_OPT=$(COVERAGE_OPT)"
	@echo "TOOLS_DIR=$(TOOLS_DIR)"
	@echo "GOBUILD=$(GOBUILD)"
	@echo "ENV=$(ENV)"

.PHONY: precommit
precommit: lint test coverage 

.PHONY: lint
lint: $(TOOLS_DIR)/golangci-lint
	$(TOOLS_DIR)/golangci-lint run $(LINT_RUN_OPTS)

.PHONY: coverage
coverage:
	$(GOTEST) $(GOTEST_OPT) $(GOTEST_COVERAGE_OPT) $(PKG_SPEC)
	go tool cover -html=coverage.txt -o coverage.html

.PHONY: test
test:
	$(GOTEST) $(GOTEST_OPT) $(PKG_SPEC)