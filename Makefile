default: build

build:
	go build ./cmd/duplicates-checker

test:
	go test -race ./... -count=1