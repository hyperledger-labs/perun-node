// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
//     https://github.com/direct-state-transfer/dst-go/NOTICE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// build implements commands that will be called from the continuous integration scripts.
// The commands are also used by makefile to build the project on systems where setting up
// a go development environment or using existing one for this project is not desirable.
//
//Available commands are
//	fetchDependencies		-- fetches all needed dependencies
//	install				-- fetch dependencies and installs the dst-go software.
//	ciInstall			-- installs the dst-go software
//	lint				-- runs meta-linter with configuration at linterConfig.json.
//	test				-- fetch dependencies and runs unit tests, all flags of "go test" command can be used.
//	ciTest				-- runs unit tests, all flags of "go test" command can be used.
//	runWalkthrough			-- fetch dependencies, builds and runs the walkthrough. Specify "simulated" or "real" backend.
//	ciRunWalkthrough		-- builds and runs the walkthrough. Specify "simulated" or "real" backend.
package main
