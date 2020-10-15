// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
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

package main

import "github.com/spf13/pflag"

// areAllFlagsSpecified returns true if all of the flags were specified
// invoking the command to which the passed flagset was attached to.
func areAllFlagsSpecified(fs *pflag.FlagSet, flags ...string) bool {
	for i := range flags {
		if !fs.Changed(flags[i]) {
			return false
		}
	}
	return true
}
