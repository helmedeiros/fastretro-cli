.PHONY: build test test-race cover cover-html clean lint fmt-check check

BINARY := fastretro
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/fastretro

test:
	go test ./... -v

test-race:
	go test ./... -race -v

cover:
	go test ./... -coverprofile=coverage.out -race
	go tool cover -func=coverage.out

cover-html:
	go test ./... -coverprofile=coverage.out -race
	go tool cover -html=coverage.out -o coverage.html

fmt-check:
	@test -z "$$(gofmt -l ./cmd ./internal)" || (echo "gofmt failed:"; gofmt -l ./cmd ./internal; exit 1)

lint:
	go vet ./...
	golangci-lint run ./...

tidy:
	go mod tidy
	go mod verify

check: fmt-check lint test-race build
	@echo "All quality gates passed."

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html
