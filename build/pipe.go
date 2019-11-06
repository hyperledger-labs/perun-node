// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
//     https://github.com/direct-state-transfer/dst-go
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
	"bytes"
	"os/exec"
)

func pipe(CommandsToExecute ...*exec.Cmd) (outputBuff bytes.Buffer, err error) {

	stderrBuff, stdoutBuff, stdinBuff := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}

	for _, cmd := range CommandsToExecute {

		stdoutBuff = &bytes.Buffer{}

		cmd.Stdin = stdinBuff
		cmd.Stderr = stderrBuff
		cmd.Stdout = stdoutBuff

		err = cmd.Run()

		if err != nil {
			return *stderrBuff, err
		}

		stdinBuff = stdoutBuff
	}

	return *stdoutBuff, nil
}
