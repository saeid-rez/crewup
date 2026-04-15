VERSION ?= dev
BINARY  := crewup
LDFLAGS := -ldflags "-X github.com/saeid-rez/crewup/cmd.Version=$(VERSION)"

.PHONY: build run clean release

## build: compile for current platform
build:
	go build $(LDFLAGS) -o $(BINARY) .

## run: build and run immediately
run: build
	./$(BINARY)

## install: install to $GOPATH/bin
install:
	go install $(LDFLAGS) .

## clean: remove built binaries
clean:
	rm -f $(BINARY)

## release: cross-compile for all platforms (requires goreleaser)
release:
	goreleaser release --snapshot --clean

## release-dry: dry run of goreleaser
release-dry:
	goreleaser release --snapshot --clean --skip=publish

## tidy: tidy go modules
tidy:
	go mod tidy
