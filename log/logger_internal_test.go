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

package log

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewLoggerWithField(t *testing.T) {
	t.Run("happy_without_init", func(t *testing.T) {
		setCleanup(t)
		logger = nil

		var l Logger
		assert.NotPanics(t, func() {
			l = NewLoggerWithField(logrus.Fields{"testkey1": "testval", "testkey2": "testval"})
		})
		require.NotNil(t, l)
		assert.Equal(t, logrus.DebugLevel, logger.Level)
		assert.Equal(t, logrus.DebugLevel, l.(*logrus.Entry).Logger.Level)
	})

	t.Run("happy_with_init_stdout", func(t *testing.T) {
		setCleanup(t)
		err := InitLogger("error", "")
		require.NoError(t, err)

		var l Logger
		assert.NotPanics(t, func() {
			l = NewLoggerWithField("testkey", "testval")
		})
		require.NotNil(t, l)
		assert.Equal(t, logrus.ErrorLevel, logger.Level)
		assert.Equal(t, logrus.ErrorLevel, l.(*logrus.Entry).Logger.Level)
	})
	t.Run("happy_with_init_file", func(t *testing.T) {
		setCleanup(t)
		tempFile, err := ioutil.TempFile("", "")
		require.NoError(t, err)
		require.NoError(t, tempFile.Close())
		t.Cleanup(func() {
			err = os.Remove(tempFile.Name())
			if err != nil {
				t.Log("Removing temp file created for test -", tempFile.Name())
			}
		})

		err = InitLogger("error", tempFile.Name())
		require.NoError(t, err)

		var l Logger
		assert.NotPanics(t, func() {
			l = NewLoggerWithField("testkey", "testval")
		})
		require.NotNil(t, l)
		assert.Equal(t, logrus.ErrorLevel, logger.Level)
		assert.Equal(t, logrus.ErrorLevel, l.(*logrus.Entry).Logger.Level)
	})
}

func Test_InitLogger(t *testing.T) {
	t.Run("err_multiple_init", func(t *testing.T) {
		setCleanup(t)
		err1 := InitLogger("error", "")
		require.NoError(t, err1)
		err2 := InitLogger("info", "")
		require.Error(t, err2)
		t.Log(err2)

		require.NotNil(t, logger)
		assert.Equal(t, logrus.ErrorLevel, logger.Level)
	})

	t.Run("err_invalid_level", func(t *testing.T) {
		setCleanup(t)
		err := InitLogger("invalid-level", "")
		require.Error(t, err)
		t.Log(err)

		require.Nil(t, logger)
	})

	t.Run("err_setting_up_file", func(t *testing.T) {
		setCleanup(t)
		tempDir, err := ioutil.TempDir("", "")
		require.NoError(t, err)
		logFile := filepath.Join(tempDir, "1")
		f, err := os.OpenFile(logFile, os.O_RDONLY|os.O_CREATE, 0o000)
		require.NoError(t, err)
		require.NoError(t, f.Close())
		t.Cleanup(func() {
			err = os.RemoveAll(tempDir)
			if err != nil {
				t.Log("Removing temp dir created for test -", tempDir)
			}
		})

		err = InitLogger("error", logFile)
		require.Error(t, err)
		t.Log(err)

		require.Nil(t, logger)
	})
}

// setCleanup backs up the original value of package level logger instance and
// registers a callback to test cleanup to restore it after the test.
func setCleanup(t *testing.T) {
	oldLogger := logger
	t.Cleanup(func() {
		logger = oldLogger
	})
}
