SOURCEDIR = .
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY = drunkenfall

VERSION = $(shell git describe --dirty --always --tags)
BUILDTIME = `date +%FT%T%z` # ISO-8601

LDFLAGS = -ldflags "-X $(BINARY).version=${VERSION} -X $(BINARY).buildtime=${BUILDTIME}"

LINTER_ARGS = -j 4 --enable-gc --disable=errcheck --deadline=10m --tests

.DEFAULT_GOAL: all

all: clean download npm test $(BINARY)

check: test lint

.PHONY: clean
clean:
	rm -f $(BINARY)

.PHONY: clobber
clobber: clean
	rm -rf js/node_modules

$(BINARY): download $(SOURCES)
	go build -v ${LDFLAGS} -o ${BINARY}

.PHONY: download
download:
	go get -t -d -v ./...

.PHONY: install
install:
	go install -v ${LDFLAGS} ./...

.PHONY: install-linter
install-linter:
	go get -v -u github.com/alecthomas/gometalinter
	gometalinter --install

.PHONY: test
test:
	go test -v -coverprofile=cover.out

.PHONY: race
race:
	go test -race -v

.PHONY: lint
lint: install-linter
	gometalinter $(LINTER_ARGS) $(SOURCEDIR)

npm: js/package.json
	cd js; npm install

.PHONY: npm
npm-start: npm
	cd js; npm run dev

$(BINARY)-start: $(BINARY)
	./$(BINARY)

.PHONY: nginx-start
nginx-start:
	mkdir -p logs
	sudo nginx -p . -c conf/nginx.conf # TODO: Make sure we can run this without sudo

# Listing all the make targets; http://stackoverflow.com/a/26339924/983746
.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'
