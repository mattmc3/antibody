export GO111MODULE := "on"

SOURCE_FILES := "./..."
TEST_PATTERN := "."
TEST_OPTIONS := "-v"

# Default target
default: build

# Run command in dev container
container CMD:
    podman run --rm -v "$PWD:/workspace:z" -w /workspace antibody-dev {{CMD}}

# Build dev container image
build-container:
    podman build -t antibody-dev .

# Run tests in container
test-container:
    just container "go test -v ./..."

# Build in container
build-bin-container:
    just container "go build"

# Run CI checks in container
ci-container:
    just container "go build && go test -v ./... && golangci-lint run ./..."

# Run all the tests
test:
    go test {{TEST_OPTIONS}} -failfast -race -coverpkg=./... -covermode=atomic -coverprofile=coverage.txt {{SOURCE_FILES}} -run {{TEST_PATTERN}} -timeout=2m

# Run all the tests and opens the coverage report
cover: test
    go tool cover -html=coverage.txt

# Run all the linters
lint:
    golangci-lint run ./...

# Run all the tests and code checks
ci: build test lint

# Build a beta version
build:
    go build

# Format all go files
fmt:
    go fmt ./...

# Build release binaries for all platforms (GitHub Actions only)
build-release VERSION:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p dist
    LDFLAGS="-X main.version={{VERSION}} -s -w"

    echo "Building binaries..."
    for target in \
        "linux/amd64" \
        "linux/arm64" \
        "darwin/amd64" \
        "darwin/arm64" \
        "freebsd/amd64"
    do
        os="${target%/*}"
        arch="${target#*/}"
        output="dist/antibody_${os}_${arch}"
        echo "  $os/$arch"
        GOOS=$os GOARCH=$arch go build -ldflags "$LDFLAGS" -o "$output"
    done

    echo "Creating archives..."
    cd dist
    for binary in antibody_*; do
        tar czf "${binary}.tar.gz" "$binary"
    done

    echo "Creating checksums..."
    sha256sum antibody_*.tar.gz > "antibody_{{VERSION}}_checksums.txt"

    echo "Creating packages..."
    for arch in amd64 arm64; do
        cat > nfpm.yaml <<EOF
    name: antibody
    arch: $arch
    platform: linux
    version: {{VERSION}}
    maintainer: mattmc3
    description: The fastest shell plugin manager
    homepage: https://github.com/mattmc3/antibody
    license: MIT
    contents:
      - src: antibody_linux_$arch
        dst: /usr/bin/antibody
    EOF
        nfpm pkg --packager deb --target "antibody_{{VERSION}}_linux_${arch}.deb"
        nfpm pkg --packager rpm --target "antibody_{{VERSION}}_linux_${arch}.rpm"
    done
    cd ..

    echo "Release artifacts ready in dist/"

# Create a release tag (strips -dev from current version)
tag:
    #!/usr/bin/env bash
    VERSION=$(grep '^current_version' .bumpversion.cfg | cut -d' ' -f3)
    echo "Creating release v$VERSION"
    echo "Make sure CHANGELOG.md is updated!"
    read -p "Press enter to continue..."
    git tag "v$VERSION"
    git push --tags

# Bump version after release (e.g., just bump patch)
bump PART='patch':
    bumpversion {{PART}}
    git push

# Prepare changelog for next release
changelog:
    #!/usr/bin/env bash
    VERSION=$(grep '^current_version' .bumpversion.cfg | cut -d' ' -f3)
    DATE=$(date +%Y-%m-%d)
    # Update [Unreleased] to [VERSION] - DATE
    sed -i.bak "s/## \[Unreleased\]/## [Unreleased]\n\n## [$VERSION] - $DATE/" CHANGELOG.md
    rm CHANGELOG.md.bak
    echo "Updated CHANGELOG.md for v$VERSION"
    echo "Edit it to add your changes, then run: just tag"
