VERSION := $(shell git describe --tags --always)
GIT_COMMIT_ID := $(shell git rev-parse HEAD)
GOPERUN_VERSION := $(shell grep "go-perun" go.mod | cut -d "v" -f2)

NODE_PKG := ./cmd/perunnode
NODE_BIN := perunnode

LDFLAGS=-ldflags "-X 'main.version=$(VERSION)' -X 'main.gitCommitID=$(GIT_COMMIT_ID)' -X 'main.goperunVersion=$(GOPERUN_VERSION)'"

build:
	go build $(LDFLAGS) $(NODE_PKG)

clean:
	rm -rf perunnode node.yaml alice bob
