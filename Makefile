.PHONY: fmt vet test build check

fmt:
	test -z "$$(gofmt -l .)"

vet:
	go vet ./...

test:
	go test -race ./...

build:
	go build ./cmd/ttft-itl-benchmark

check: fmt vet test build
