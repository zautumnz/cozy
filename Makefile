PREFIX ?= /usr/local
VERSION := $(shell git describe --tags 2>/dev/null)

build:
	@go build -ldflags "-X main.COZY_VERSION=$(VERSION)"

install:
	@mkdir -p $(PREFIX)/bin
	@cp -f cozy $(PREFIX)/bin/cozy
	@chmod 755 $(PREFIX)/bin/cozy

fmt:
	go fmt ./...

clean:
	@rm -f cozy coverage.out

test:
	@go test ./...

cover:
	@go test -coverprofile=coverage.out ./...

coverage:
	@go tool cover -html=coverage.out

count:
	@cloc --exclude-dir=x,.git,.github --read-lang-def=editor/cozy.cloc .

lint:
	@go fmt ./...
	@go vet ./...
	@staticcheck ./...

tags:
	@ctags --exclude=x -R .

.PHONY: clean install tags todo
