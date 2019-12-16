# You can set these variables from the command line.
BUILDOPTS =

# Internal variables
__BUILDHELP = (go doc ./build \
              | grep -E '^    [[:alnum:]_-]+.*?--.*|^         .*')
__TARGETS = $(shell $(__BUILDHELP) \
            | awk 'BEGIN {FS=" -- |^      *?"}; $$1 != "" {print $$1}')

# color definitions for help output
__BLUE = \033[1;34m
__NC   = \033[0m

# Put it first so that "make" without argument is like "make help".
help:               ## display this help text
	@echo -n "Usage: make $(__BLUE)target$(__NC) [BUILDOPTS='options']"
	@echo -n "\n\nThe available targets are:\n"
	@grep -E '^[[:alnum:]_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; \
	       {printf "  $(__BLUE)%-17s$(__NC) %s\n", $$1, $$2}'
	@$(__BUILDHELP) \
	| awk 'BEGIN {FS=" -- |^      *?"}; \
	       {gsub(/ /, "", $$1); \
	        printf "  $(__BLUE)%-17s$(__NC) %s\n", $$1, $$2}'
	@echo -n "\nExample: make test BUILDOPTS='-short'\n"

.PHONY: help Makefile

# $(O) is meant as a shortcut for $(BUILDOPTS).
$(__TARGETS): Makefile
	go build -o build/build ./build
	build/setEnv.sh build/build $@ $(BUILDOPTS) $(0)
	rm build/build
	@echo "Done building."
