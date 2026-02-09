BINARY := aitop
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/isaacaudet/aitop/cmd.Version=$(VERSION)"

.PHONY: build test clean install

build:
	CGO_ENABLED=1 go build $(LDFLAGS) -o $(BINARY) .

test:
	CGO_ENABLED=1 go test ./... -v

clean:
	go clean
	$(RM) $(BINARY)

install:
	CGO_ENABLED=1 go install $(LDFLAGS) .
