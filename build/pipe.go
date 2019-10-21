package main

import (
	"bytes"
	"os/exec"
)

func pipe(CommandsToExecute ...*exec.Cmd) (outputBuff bytes.Buffer, err error) {

	stderrBuff, stdoutBuff, stdinBuff := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}

	for _, cmd := range CommandsToExecute {
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
