SOURCEDIR = .
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY = drunkenfall

VERSION = $(shell git describe --always --tags)
BUILDTIME = `date +%FT%T%z` # ISO-8601

LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.buildtime=${BUILDTIME}"

export GOPATH := $(shell go env GOPATH)
# export PATH := $(GOPATH)/bin:$(PATH)
.DEFAULT_GOAL: all

.PHONY: clean clobber download install install-linter test cover race lint npm npm-dist caddy

all: clean download npm test race lint $(BINARY)

check: test lint

clean:
	rm -f $(BINARY)

clobber: clean
	rm -rf js/node_modules

BINARY = drunkenfall
$(BINARY): $(SOURCES)
	go build -v ${LDFLAGS} -o ${BINARY}

.PHONY: $(BINARY)-start
$(BINARY)-start: $(BINARY)
	./$(BINARY)

.PHONY: dist
dist: $(BINARY)
	cd js; npm run build

download:
	go get -t -d -v ./...

install:
	go install -v ${LDFLAGS} ./...

install-linter:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin v1.12.2

test:
	GIN_MODE=test go test -v ./towerfall

cover:
	go test -coverprofile=cover.out ./...

race:
	go test -race -v ./...

lint:
	golangci-lint run

npm: js/package.json
	cd js; npm install

npm-start:
	cd js; PORT=42002 npm run dev

npm-sass:
	cd js; npm rebuild node-sass

npm-dist: npm
	cd js; npm run build

.PHONY: vendor
vendor:
	dep ensure -v

.PHONY: docker
docker:
	docker-compose build

.PHONY: download-caddy
download-caddy:
	go get github.com/mholt/caddy/caddy
	go get github.com/caddyserver/builds
	cd $(GOPATH)/src/github.com/mholt/caddy/caddy; go run build.go

.PHONY: caddy
caddy: download-caddy
	sudo $(GOPATH)/bin/caddy

.PHONY: caddy-local
caddy-local: download-caddy
	sudo $(GOPATH)/bin/caddy -conf Caddyfile.local

.PHONY: postgres
postgres:
	docker-compose up postgres

.PHONY: psql
psql:
	@psql --host localhost --user postgres drunkenfall

DB := ./data/db.sql

.PHONY: reset-db
reset-db:
	test -n "$(DRUNKENFALL_RESET_DB)"
	psql --host localhost --user postgres -c "DROP DATABASE drunkenfall"
	psql --host localhost --user postgres -c "CREATE DATABASE drunkenfall"
	psql --host localhost --user postgres drunkenfall < $(DB)

.PHONY: reset-test-db
reset-test-db:
	psql --host localhost --user postgres -c "DROP DATABASE test_drunkenfall"
	psql --host localhost --user postgres -c "CREATE DATABASE test_drunkenfall"
	psql --host localhost --user postgres test_drunkenfall < $(DB)

.PHONY: make-test-db
make-test-db:
	pg_dump --user postgres --host localhost drunkenfall > $(DB)
