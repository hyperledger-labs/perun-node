VERSION := $(shell git describe --tags --always)
GIT_COMMIT_ID := $(shell git rev-parse HEAD)
GOPERUN_VERSION := $(shell grep "go-perun" go.mod | cut -d "v" -f2)

NODE_PKG := ./cmd/perunnode
NODE_BIN := perunnode

CLI_PKG := ./cmd/perunnodecli
CLI_BIN := perunnodecli

TUI_PKG := ./cmd/perunnodetui
TUI_BIN := perunnodetui

DEMO_DIR := demo

LDFLAGS=-ldflags "-X 'main.version=$(VERSION)' -X 'main.gitCommitID=$(GIT_COMMIT_ID)' -X 'main.goperunVersion=$(GOPERUN_VERSION)'"

install:
	go install $(LDFLAGS) $(NODE_PKG)
	go install $(CLI_PKG)
	go install $(TUI_PKG)

generate: install
	@mkdir $(DEMO_DIR)
	@cd $(DEMO_DIR) && $(NODE_BIN) generate
	@echo "Configuration files for demo generated in ./$(DEMO_DIR)"

clean:
	rm -rf demo
