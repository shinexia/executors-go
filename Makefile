.PHONY: all
all: tidy lint build test

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: build
build:
	go build -v ./...

.PHONY: test
test:
	go test -count=1 ./...
