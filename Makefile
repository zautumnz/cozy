PREFIX ?= /usr/local
CONFIG_PREFIX ?= /usr/share

build:
	@go build -ldflags "-X main.version=$(git describe --tags 2>/dev/null)"

install:
	@mkdir -p $(PREFIX)/bin
	@cp -f cozy $(PREFIX)/bin/cozy
	@chmod 755 $(PREFIX)/bin/cozy

fmt:
	go fmt ./...

clean:
	@rm -f cozy

test:
	@go test ./...

count:
	@cloc --exclude-dir=x,.git,.github --read-lang-def=editor/cozy.cloc .

lint:
	@go fmt ./...
	@go vet ./...
	@staticcheck ./...

tags:
	@ctags --exclude=x -R .

.PHONY: clean install tags todo
