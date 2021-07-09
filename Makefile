PREFIX ?= /usr/local
CONFIG_PREFIX ?= /usr/share

cozy:
	go build -ldflags "-X main.version=$(git describe --tags 2>/dev/null || echo 'master')"

install:
	mkdir -p $(PREFIX)/bin
	cp -f cozy $(PREFIX)/bin/cozy
	chmod 755 $(PREFIX)/bin/cozy

clean:
	rm -f cozy

test:
	go test ./...

count:
	cloc --exclude-dir=x --read-lang-def=editor/cozy.cloc .

lint:
	go fmt ./...
	go vet ./...
	staticcheck ./...

.PHONY: cozy clean install count lint test
