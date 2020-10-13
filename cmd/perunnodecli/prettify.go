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

import (
	"fmt"
	"reflect"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

var prettyFormatterOverrides = map[reflect.Type]interface{}{
	reflect.TypeOf(time.Duration(0)): fmt.Sprint,
}

var prettyFormatterConfig = &pretty.Config{
	Formatter: prettyFormatterOverrides,
}

// prettify returns a prettified string version of the input data.
// For structs, it includes both exported and unexported fields.
// Formatting of time strings is preserved.
func prettify(vals ...interface{}) string {
	return prettyFormatterConfig.Sprint(vals...)
}
