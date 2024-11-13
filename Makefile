.PHONY: all
all: lint test

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: test
test:
	go test -count=1 ./...
