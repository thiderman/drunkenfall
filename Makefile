SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY=drunkenfall

VERSION=$(shell git describe --dirty --always --tags)
BUILDTIME=`date +%FT%T%z` # ISO-8601

LDFLAGS=-ldflags "-X $(BINARY).version=${VERSION} -X $(BINARY).buildtime=${BUILDTIME}"

.DEFAULT_GOAL: all
.PHONY: download install test clean npm npm-start $(BINARY)-start nginx-start

all: clean download npm test $(BINARY)

$(BINARY): download $(SOURCES)
	go build -v ${LDFLAGS} -o ${BINARY}

download:
	go get -t -d -v ./...

install:
	go install -v ${LDFLAGS} ./...

test:
	go test -v

clean:
	rm -f $(BINARY)

npm:
	cd js; npm install

npm-start: npm
	cd js; npm run dev

$(BINARY)-start: $(BINARY)
	./$(BINARY)

nginx-start:
	mkdir -p logs
	sudo nginx -p . -c conf/nginx.conf # TODO: Make sure we can run this without sudo
