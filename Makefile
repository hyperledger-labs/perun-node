
.PHONY: dst

fetchDependencies:
	cd build && go build && cd ..
	build/setEnv.sh build/build fetchDependencies
	rm build/build
	@echo "Done fetching."

install:
	cd build && go build && cd ..
	build/setEnv.sh build/build install
	rm build/build
	@echo "Done building."

ciInstall:
	cd build && go build && cd ..
	build/setEnv.sh build/build ciInstall
	rm build/build
	@echo "Done ci building."

test:
	cd build && go build && cd ..
	build/setEnv.sh build/build test $(testOpts)
	rm build/build
	@echo "Test Completed."

ciTest:
	cd build && go build && cd ..
	build/setEnv.sh build/build ciTest $(testOpts)
	rm build/build
	@echo "Test Completed."


runWalkthrough: 
	cd build && go build && cd ..
	build/setEnv.sh build/build runWalkthrough $(walkthroughOpts)
	rm build/build
	@echo "walkthrough completed."

ciRunWalkthrough: 
	cd build && go build && cd ..
	build/setEnv.sh build/build ciRunWalkthrough $(walkthroughOpts)
	rm build/build
	@echo "justrunwalkthrough completed."

lint: 
	cd build && go build && cd ..
	build/setEnv.sh build/build lint
	rm build/build
	@echo "Performed Lint."


target: install
