
.PHONY: dst

install:
	cd build && go build && cd ..
	build/setEnv.sh build/build install
	rm build/build
	@echo "Done building."

test:
	cd build && go build && cd ..
	build/setEnv.sh build/build test $(testOpts)
	rm build/build
	@echo "Test Completed."

runWalkthrough: 
	cd build && go build && cd ..
	build/setEnv.sh build/build runWalkthrough $(walkthroughOpts)
	rm build/build
	@echo "walkthrough completed."

lint: 
	cd build && go build && cd ..
	build/setEnv.sh build/build lint
	rm build/build
	@echo "Performed Lint."


target: install
