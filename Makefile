GO             ?= go
GOOS           ?= $(shell $(GO) env GOOS)
GIT_TAG        := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

MAKEFILE       := $(realpath $(lastword $(MAKEFILE_LIST)))
SOURCES        := $(shell find . -name '*.go') $(MAKEFILE)

BUILD_FLAGS    := -a -ldflags="-s -w -X github.com/NitorCreations/tai/internal/cli.Version=$(GIT_TAG)" -trimpath

BINARY64       := tai_$(GOOS)_amd64
BINARYARM8     := tai_$(GOOS)_arm8

# https://en.wikipedia.org/wiki/Uname
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
	BINARY := $(BINARY64)
else ifeq ($(UNAME_M),amd64)
	BINARY := $(BINARY64)
else ifeq ($(UNAME_M),arm64)
	BINARY := $(BINARYARM8)
else ifeq ($(UNAME_M),aarch64)
	BINARY := $(BINARYARM8)
else
$(error Build on $(UNAME_M) is not supported, yet.)
endif

all: dist/$(BINARY)

clean:
	$(RM) -r dist

dist/$(BINARY64): $(SOURCES)
	GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o $@ ./cmd/tai/

dist/$(BINARYARM8): $(SOURCES)
	GOARCH=arm64 $(GO) build $(BUILD_FLAGS) -o $@ ./cmd/tai/

dist/tai: dist/$(BINARY) | dist
	rm -f dist/tai
	cp -f dist/$(BINARY) dist/tai

update:
	$(GO) get -u ./...
	$(GO) mod tidy

.PHONY: all update