GO=go

default: build

build:
	$(GO) build -o bin/kctxpatch cmd/kctxpatch/*.go

.PHONY: build
