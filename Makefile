VERSION := $(shell git describe --tags --always)
GIT_COMMIT_ID := $(shell git rev-parse HEAD)
GOPERUN_VERSION := $(shell grep "go-perun" go.mod | cut -d "v" -f2)

NODE_PKG := ./cmd/perunnode
NODE_BIN := perunnode

CLI_PKG := ./cmd/perunnodecli
CLI_BIN := perunnodecli

LDFLAGS=-ldflags "-X 'main.version=$(VERSION)' -X 'main.gitCommitID=$(GIT_COMMIT_ID)' -X 'main.goperunVersion=$(GOPERUN_VERSION)'"

build:
	go build $(LDFLAGS) $(NODE_PKG)
	go build $(CLI_PKG)

clean:
	rm -rf $(NODE_BIN) $(CLI_BIN) node.yaml alice bob
