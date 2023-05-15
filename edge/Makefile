PROJECT_NAME := "kronos"

all: build

.PHONY: build
build:
	./scripts/build.sh

.PHONY: clean
clean:
	rm ./cmd/kronos/kronos-*

.PHONY: dep
dep:
	go get -v -d ./...

.PHONY: test
test:
	go test ./...

.PHONY: race
race:
	go test -race ./...

.PHONY: cover
cover:
	go test ./... -coverprofile=coverage.out

.PHONY: show-cover
show-cover:
	go tool cover -html=coverage.out

.PHONY: cloc
cloc:
	gocloc .

.PHONY: lint
lint:
	 golangci-lint run ./... --timeout 5m
