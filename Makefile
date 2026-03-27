.PHONY: build test test-integration lint fmt clean yrs install

all: test lint build

build:
	go build ./...

test:
	go test -v ./...

test-integration:
	go test -v -tags=integration ./...

lint:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -f coverage.out
	rm -rf bin/

yrs:
	go build -o bin/yrs ./cmd/yrs

install:
	go install ./cmd/yrs
