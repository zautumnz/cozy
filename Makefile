PREFIX ?= /usr/local
VERSION := $(shell git describe --tags 2>/dev/null)

build:
	@go build -ldflags "-X main.COZY_VERSION=$(VERSION)"

install:
	@mkdir -p $(PREFIX)/bin
	@cp -f cozy $(PREFIX)/bin/cozy
	@chmod 755 $(PREFIX)/bin/cozy

clean:
	@rm -f cozy coverage.out

cover:
	@go test -coverprofile=coverage.out ./...

cover_open:
	@go tool cover -html=coverage.out

count:
	@cloc --exclude-dir=x,.git,.github,examples --read-lang-def=editor/cozy.cloc .

test:
	@go fmt ./...
	@go vet ./...
	@staticcheck ./...
	@go test ./...

tags:
	@ctags --exclude=x --exclude=examples --exclude=editor -R .

.PHONY: clean install tags
