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

package identity

import (
	"testing"

	"github.com/direct-state-transfer/dst-go/log"
)

func Test_FlagSet(t *testing.T) {

	flagSet := GetFlagSet()
	requiredFlags := []string{"identityLogLevel", "identityLogBackend"}

	failed := false
	t.Run("GetFlagset", func(t *testing.T) {
		for _, val := range requiredFlags {
			if nil == flagSet.Lookup(val) {
				failed = true
				t.Errorf("GetFlagSet() %s flag not defined", val)
			}
		}
	})

	if failed {
		return
	}

	t.Run("ParseFlags", func(t *testing.T) {

		sampleValues := map[string]string{
			"identityLogLevel":   "Debug",
			"identityLogBackend": "stdout",
		}
		for key, value := range sampleValues {
			err := flagSet.Set(key, value)
			if err != nil {
				t.Fatalf("Setup() error setting flag %s to value %s", key, value)
			}
		}

		var testCfg = &Config{}
		err := ParseFlags(flagSet, testCfg)
		if err != nil {
			t.Fatalf("ParseFlags() error=%v, want nil", err)
		}

	})

}

func Test_SetLogger(t *testing.T) {
	testLogger, err := log.NewLogger(log.DebugLevel, log.StdoutBackend, "identity-test")
	if err != nil {
		t.Fatalf("Setup: log.NewLogger() error=%v", err)
	}
	SetLogger(testLogger)
}
