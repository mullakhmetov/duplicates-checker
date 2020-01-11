VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo "1.0.0")
LDFLAGS := -ldflags "-X main.revision=${VERSION}"

default: build

build:
	go build ${LDFLAGS} ./cmd/duplicates-checker

test:
	go test -race ./... -count=1

lint:
	golint -set_exit_status ./...