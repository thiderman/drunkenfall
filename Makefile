SOURCEDIR = .
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY = drunkenfall

VERSION = $(shell git describe --always --tags)
BUILDTIME = `date +%FT%T%z`

LDFLAGS = -ldflags "-w -X main.version=${VERSION} -X main.buildtime=${BUILDTIME}"

export GOPATH := $(shell go env GOPATH)
.DEFAULT_GOAL: $(BINARY)

check: test lint

.PHONY: lint
lint:
	golangci-lint run --config .golangci-pedantic.yaml

# The CGO and GOOS are needed for it to statically link so that it is runnable in Docker
.PHONY: install
install:
	CGO_ENABLED=0 GOOS=linux go install -v ${LDFLAGS} ./...

BINARY = drunkenfall
$(BINARY): $(SOURCES)
	go build -v ${LDFLAGS} -o ${BINARY}

.PHONY: $(BINARY)-start
$(BINARY)-start: $(BINARY)
	./$(BINARY)

.PHONY: dist
dist: $(BINARY)
	cd js; npm run build

.PHONY: npm
npm: js/package.json
	cd js; npm install

.PHONY: npm-start
npm-start:
	cd js; PORT=42002 npm run dev

.PHONY: npm-dist
npm-dist: npm
	cd js; npm run build

.PHONY: frontend
frontend:
	echo "export const version = '$(VERSION) ($(BUILDTIME))'" > js/src/version.js
	docker-compose build frontend
	rm -rf dist/
	docker cp $(shell docker create drunkenfall_frontend:latest):/dist ./dist
	docker rm $(shell docker container ls -a | grep frontend | cut -f1 -d" ")

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
caddy:
	sudo $(GOPATH)/bin/caddy

.PHONY: caddy-local
caddy-local: download-caddy
	sudo $(GOPATH)/bin/caddy -conf Caddyfile.local

.PHONY: postgres
postgres:
	docker-compose up postgres

DB := ./data/db.sql
DBPARAMS := --host localhost --user postgres

.PHONY: reset-db
reset-db:
	test -n "$(DRUNKENFALL_RESET_DB)"
	psql $(DBPARAMS) -c "DROP DATABASE drunkenfall"
	psql $(DBPARAMS) -c "CREATE DATABASE drunkenfall"
	psql $(DBPARAMS) drunkenfall < $(DB)

.PHONY: reset-test-db
reset-test-db:
	psql $(DBPARAMS) -c "DROP DATABASE test_drunkenfall"
	psql $(DBPARAMS) -c "CREATE DATABASE test_drunkenfall"
	psql $(DBPARAMS) test_drunkenfall < $(DB)

.PHONY: make-test-db
make-test-db:
	pg_dump $(DBPARAMS) drunkenfall > $(DB)
