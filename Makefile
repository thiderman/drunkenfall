SOURCEDIR = .
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY = drunkenfall

VERSION = $(shell git describe --dirty --always --tags)
BUILDTIME = `date +%FT%T%z` # ISO-8601

LDFLAGS = -ldflags "-X main.version=${VERSION} -X main.buildtime=${BUILDTIME}"

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

.PHONY: clean clobber download install install-linter test cover race lint npm npm-dist nginx-start

all: clean download npm test race lint $(BINARY)

check: test lint

clean:
	rm -f $(BINARY)

clobber: clean
	rm -rf js/node_modules

$(BINARY): download $(SOURCES)
	go build -v ${LDFLAGS} -o ${BINARY}

download:
	go get -t -d -v ./...

install:
	go install -v ${LDFLAGS} ./...

install-linter:
	go get -v -u github.com/alecthomas/gometalinter
	gometalinter --install

test:
	go test -v

cover:
	go test -coverprofile=cover.out

race:
	go test -race -v

lint: install-linter
	gometalinter $(LINTER_ARGS) $(SOURCEDIR)

npm: js/package.json
	cd js; npm install

npm-start: npm
	cd js; npm run dev

npm-dist: npm
	cd js; npm run build

$(BINARY)-start: $(BINARY)
	./$(BINARY)

nginx-start:
	mkdir -p logs
	sudo nginx -p . -c conf/nginx.conf
