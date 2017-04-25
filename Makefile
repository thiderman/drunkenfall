SOURCEDIR = .
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY = drunkenfall

VERSION = $(shell git describe --dirty --always --tags)
BUILDTIME = `date +%FT%T%z` # ISO-8601

LDFLAGS = -ldflags "-X $(BINARY).version=${VERSION} -X $(BINARY).buildtime=${BUILDTIME}"

export GOPATH := $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)
# gotype is disabled since it seems pointless and also produces 250 errors
# about not finding dependencies that definitely exists.
LINTER_ARGS = -j 4 \
  --enable-gc \
  --enable=gofmt \
  --enable=misspell \
  --enable=unparam \
  --enable=unused \
  --disable=errcheck \
  --disable=gotype \
  --deadline=10m \
  --tests

.DEFAULT_GOAL: all

all: clean download npm test race lint $(BINARY)

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
	go test

.PHONY: cover
cover:
	go test -coverprofile=cover.out

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

.PHONY: npm-dist
npm-dist: npm
	cd js; npm run build

$(BINARY)-start: $(BINARY)
	./$(BINARY)

.PHONY: nginx-start
nginx-start:
	mkdir -p logs
	sudo nginx -p . -c conf/nginx.conf # TODO: Make sure we can run this without sudo
