export PATH := "./bin:" + env_var('PATH')
export GO111MODULE := "on"

SOURCE_FILES := "./..."
TEST_PATTERN := "."
TEST_OPTIONS := "-v"

# Default target
default: build

# Install all the build and lint dependencies
setup:
    curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh
    go mod download

# Run all the tests
test:
    go test {{TEST_OPTIONS}} -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt {{SOURCE_FILES}} -run {{TEST_PATTERN}} -timeout=2m

# Run all the tests and opens the coverage report
cover: test
    go tool cover -html=coverage.txt

# Run all the linters
lint:
    ./bin/golangci-lint run --disable godox --disable wsl --disable gomnd --disable testpackage --disable gofumpt --disable godot --disable nlreturn --enable-all ./...

# Run all the tests and code checks
ci: build test lint

# Build a beta version
build:
    go build

# gofmt and goimports all go files
fmt:
    find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done
