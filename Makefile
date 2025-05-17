.PHONY: fmt vet lint test build

fmt:
    go fmt ./...

vet:
    go vet ./...

lint:
    golangci-lint run

test:
    go test ./... -cover

build:
    go build -o bin/myapp ./cmd/myapp
