default: build

build:
	go build -v ./...

install: build
	go install -v ./...

# See https://golangci-lint.run/
lint:
	golangci-lint run

generate:
	go generate ./...

fmt:
	gofmt -s -w -e .

.PHONY: build install lint generate fmt