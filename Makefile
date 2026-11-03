# =============================================================================
#  ext-tags — build, test and release automation
# -----------------------------------------------------------------------------
#  Quick start:
#    make             show available targets
#    make build       build bin/ext-tags[.exe]
#    make check       format check + go vet + golangci-lint + tests
#
#  Requirements:
#    * Go 1.24+            https://go.dev/dl/
#    * GNU Make 4.x        Windows: scoop install make
#    * golangci-lint v2    Windows: scoop install golangci-lint
#
#  Windows notes:
#    * Recipes are executed by cmd.exe: Git Bash cannot be started reliably
#      in every environment, and cmd.exe is always present.
#    * git is called with -c safe.directory=* because the project may live on
#      a network share / VM shared folder, where git refuses to trust the
#      repository owner ("dubious ownership").
#    * Binaries are stamped with -buildvcs=false for the same reason; the
#      version, commit and build date are injected through -ldflags instead.
# =============================================================================

ifeq ($(OS),Windows_NT)
  SHELL       := cmd.exe
  .SHELLFLAGS := /c
  EXE         := .exe
  NULL        := NUL
  MKDIR        = powershell -NoProfile -ExecutionPolicy Bypass -Command "New-Item -ItemType Directory -Force -Path '$(1)' | Out-Null"
  RMDIR        = powershell -NoProfile -ExecutionPolicy Bypass -Command "if (Test-Path -LiteralPath '$(1)') { Remove-Item -LiteralPath '$(1)' -Recurse -Force }"
  RMFILE       = powershell -NoProfile -ExecutionPolicy Bypass -Command "if (Test-Path -LiteralPath '$(1)') { Remove-Item -LiteralPath '$(1)' -Force }"
  ENV          = set "GOOS=$(1)" && set "GOARCH=$(2)" &&
else
  EXE         :=
  NULL        := /dev/null
  MKDIR        = mkdir -p $(1)
  RMDIR        = rm -rf $(1)
  RMFILE       = rm -f $(1)
  ENV          = GOOS=$(1) GOARCH=$(2)
endif

APP      := ext-tags
PKG      := ./cmd/$(APP)
BIN_DIR  := bin
DIST_DIR := dist
BINARY   := $(BIN_DIR)/$(APP)$(EXE)

GO       ?= go
GIT      := git -c "safe.directory=*"
GOLANGCI_LINT ?= golangci-lint
GOLANGCI_VERSION ?= v2.15.0

# cmd.exe eats a lone % in command lines, so the date format is escaped on Windows.
ifeq ($(OS),Windows_NT)
  DATE_FMT := %%cI
else
  DATE_FMT := %cI
endif

# Version stamp: tag or short hash, plus -dirty when the tree is modified.
VERSION    ?= $(shell $(GIT) describe --tags --always --dirty 2>$(NULL))
COMMIT     ?= $(shell $(GIT) rev-parse --short HEAD 2>$(NULL))
BUILD_DATE ?= $(shell $(GIT) log -1 --format=$(DATE_FMT) 2>$(NULL))

ifeq ($(strip $(VERSION)),)
  VERSION := 0.0.0-dev
endif
ifeq ($(strip $(COMMIT)),)
  COMMIT := none
endif
ifeq ($(strip $(BUILD_DATE)),)
  BUILD_DATE := unknown
endif

BUILD_FLAGS := -trimpath -buildvcs=false
LDFLAGS     := -s -w -X external-tags/internal/version.Version=$(VERSION) -X external-tags/internal/version.Commit=$(COMMIT) -X external-tags/internal/version.Date=$(BUILD_DATE)

.DEFAULT_GOAL := help

.PHONY: help setup tools build run install dist test test-race cover cover-html \
        bench vet lint lint-fix fmt fmt-check tidy tidy-check check ci clean version tag

## help: list available targets
help:
	@echo ext-tags build automation
	@echo ------------------------------------------------------------------
	@echo   make build         build $(BINARY) for the current platform
	@echo   make run           run the CLI, e.g. make run ARGS="track.flac"
	@echo   make install       install the CLI into GOBIN
	@echo   make dist          cross-compile release binaries into $(DIST_DIR)/
	@echo ------------------------------------------------------------------
	@echo   make test          run unit tests
	@echo   make cover         run tests with a coverage summary
	@echo   make cover-html    render coverage.html
	@echo   make bench         run benchmarks
	@echo   make test-race     run tests with the race detector (needs cgo)
	@echo ------------------------------------------------------------------
	@echo   make fmt           format sources
	@echo   make fmt-check     fail if sources are not formatted
	@echo   make vet           run go vet
	@echo   make lint          run golangci-lint
	@echo   make lint-fix      run golangci-lint with --fix
	@echo ------------------------------------------------------------------
	@echo   make tidy          run go mod tidy
	@echo   make check         fmt-check, vet, lint and tests
	@echo   make ci            tidy-check followed by check
	@echo   make setup         verify toolchain, fetch modules, install tools
	@echo   make clean         remove build artifacts
	@echo   make version       print the version stamp
	@echo   make tag           tag the current commit, e.g. make tag TAG=v0.4.0
	@echo ------------------------------------------------------------------

## setup: check the toolchain, download modules and install dev tools
setup: tools
	@$(GO) version
	@$(GO) mod download
	@echo Setup complete.

## tools: install golangci-lint if it is missing
tools:
ifeq ($(OS),Windows_NT)
	@where $(GOLANGCI_LINT) >$(NULL) 2>&1 || scoop install golangci-lint || $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
else
	@command -v $(GOLANGCI_LINT) >$(NULL) 2>&1 || $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
endif

## build: build the CLI for the current platform
build:
	@$(call MKDIR,$(BIN_DIR))
	@echo Building $(BINARY) [$(VERSION)]
	@$(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)
	@echo Built  $(BINARY)

## run: run the CLI; pass paths through ARGS
run:
	@$(GO) run $(PKG) $(ARGS)

## install: install the CLI into GOBIN
install:
	@$(GO) install $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" $(PKG)

## dist: cross-compile release binaries into dist/
dist:
	@$(call MKDIR,$(DIST_DIR))
	@$(call ENV,windows,amd64) $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)_windows_amd64.exe $(PKG)
	@$(call ENV,windows,arm64) $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)_windows_arm64.exe $(PKG)
	@$(call ENV,linux,amd64) $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)_linux_amd64 $(PKG)
	@$(call ENV,linux,arm64) $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)_linux_arm64 $(PKG)
	@$(call ENV,darwin,amd64) $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)_darwin_amd64 $(PKG)
	@$(call ENV,darwin,arm64) $(GO) build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)_darwin_arm64 $(PKG)
	@echo Built release binaries in $(DIST_DIR)/

## test: run unit tests
test:
	@$(GO) test -count=1 ./...

## test-race: run unit tests with the race detector
test-race:
ifeq ($(OS),Windows_NT)
	@echo Skipped: the race detector requires cgo and a C compiler (gcc).
else
	CGO_ENABLED=1 $(GO) test -race -count=1 ./...
endif

## cover: run tests and print the coverage summary
cover:
	@$(GO) test -count=1 -covermode=atomic -coverprofile=coverage.out ./...
	@$(GO) tool cover -func=coverage.out

## cover-html: render coverage.out into coverage.html
cover-html: cover
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo Wrote coverage.html

## bench: run benchmarks (no benchmarks are defined yet)
bench:
	@$(GO) test -run=NoMatch -bench=. -benchmem ./...

## vet: run go vet
vet:
	@$(GO) vet ./...

## lint: run golangci-lint
lint:
	@$(GOLANGCI_LINT) run ./...

## lint-fix: run golangci-lint and apply automatic fixes
lint-fix:
	@$(GOLANGCI_LINT) run --fix ./...

## fmt: format sources with gofmt and goimports
fmt:
	@$(GOLANGCI_LINT) fmt

## fmt-check: fail when sources are not formatted
fmt-check:
	@$(GOLANGCI_LINT) fmt --diff

## tidy: run go mod tidy
tidy:
	@$(GO) mod tidy

## tidy-check: fail when go.mod/go.sum are not tidy
tidy-check:
	@$(GO) mod tidy
	@$(GIT) diff --exit-code -- go.mod go.sum

## check: format check, vet, lint and tests
check: fmt-check vet lint test
	@echo All checks passed.

## ci: full continuous-integration pipeline
ci: tidy-check check

## clean: remove build artifacts
clean:
	@$(call RMFILE,$(BINARY))
	@$(call RMDIR,$(DIST_DIR))
	@$(call RMFILE,coverage.out)
	@$(call RMFILE,coverage.html)
	@echo Cleaned.

## tag: create an annotated release tag, e.g. make tag TAG=v0.4.0
tag:
	@$(if $(TAG),,$(error TAG is required, e.g. make tag TAG=v0.4.0))
	@$(if $(shell $(GIT) status --porcelain),$(error working tree is dirty, commit the release first))
	@$(GIT) tag -a $(TAG) -m "Release $(TAG)"
	@echo Created tag $(TAG)
## version: print the version stamp
version:
	@echo $(VERSION) (commit $(COMMIT), built $(BUILD_DATE))
