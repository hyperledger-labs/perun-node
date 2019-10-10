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

package log

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_logger_methods_panic(t *testing.T) {

	var logger *Logger
	tests := []struct {
		name       string
		methodCall func()
	}{
		{"Error", func() {
			logger.Error("test-message")
		}},
		{"Info", func() {
			logger.Info("test-message")
		}},
		{"Debug", func() {
			logger.Debug("test-message")
		}},
		{"GetLogLevel", func() {
			logger.GetLogLevel()
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !assert.NotPanics(t, tt.methodCall) {
				t.Errorf("panic when *logger is uninitialized")
			}
		})
	}
}

func Test_loggerInterface_methods_panic(t *testing.T) {

	var loggerInterface = Logger{}
	tests := []struct {
		name       string
		methodCall func()
	}{
		{"Error", func() {
			loggerInterface.Error("test-message")
		}},
		{"Info", func() {
			loggerInterface.Info("test-message")
		}},
		{"Debug", func() {
			loggerInterface.Debug("test-message")
		}},
		{"GetLogLevel", func() {
			loggerInterface.GetLogLevel()
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !assert.NotPanics(t, tt.methodCall) {
				t.Errorf("panic when *LoggerInterface is set to *Logger_uninitialised")
			}
		})
	}

	var loggerNilInterface LoggerInterface
	tests = []struct {
		name       string
		methodCall func()
	}{
		{"Error", func() {
			loggerNilInterface.Error("test-message")
		}},
		{"Info", func() {
			loggerNilInterface.Info("test-message")
		}},
		{"Debug", func() {
			loggerNilInterface.Debug("test-message")
		}},
		{"GetLogLevel", func() {
			loggerNilInterface.GetLogLevel()
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !assert.Panics(t, tt.methodCall) {
				t.Errorf("does not panic when *LoggerInterface uninitialised")
			}
		})
	}

}
func Test_Config_GetLogLevel(t *testing.T) {
	type args struct {
		logLevel Level
		backend  io.Writer
	}
	tests := []struct {
		name         string
		args         args
		wantLogLevel Level
	}{
		{"Debug", args{DebugLevel, StdoutBackend}, DebugLevel},
		{"Info", args{InfoLevel, StdoutBackend}, InfoLevel},
		{"Error", args{ErrorLevel, StdoutBackend}, ErrorLevel},

		{"Debug", args{DebugLevel, StderrBackend}, DebugLevel},
		{"Info", args{InfoLevel, StderrBackend}, InfoLevel},
		{"Error", args{ErrorLevel, StderrBackend}, ErrorLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logConfig := Config{
				Level:   tt.args.logLevel,
				Backend: tt.args.backend}
			gotLogLevel := logConfig.GetLogLevel()
			if tt.wantLogLevel != gotLogLevel {
				t.Errorf("Config GetLogLevel() got %v, want %v", gotLogLevel, tt.wantLogLevel)
			}

		})
	}
}

func Test_Config_SetLogLevel(t *testing.T) {
	type args struct {
		logLevel Level
	}
	tests := []struct {
		name         string
		args         args
		wantLogLevel Level
		wantErr      bool
	}{
		{"Debug", args{DebugLevel}, DebugLevel, false},
		{"Info", args{InfoLevel}, InfoLevel, false},
		{"Error", args{ErrorLevel}, ErrorLevel, false},

		{"warn", args{Level(logrus.WarnLevel)}, InfoLevel, true},
		{"Fatal", args{Level(logrus.FatalLevel)}, DebugLevel, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var logConfig = new(Config)
			err := logConfig.SetLogLevel(tt.args.logLevel)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Config SetLogLevel() got error %v, want nil", err)
			}
		})
	}
}

func Test_isLogLevelValid(t *testing.T) {
	type args struct {
		logLevel Level
	}
	tests := []struct {
		name      string
		args      args
		wantValid bool
	}{
		{"Debug", args{DebugLevel}, true},
		{"Info", args{InfoLevel}, true},
		{"Error", args{ErrorLevel}, true},

		{"Warn", args{Level(logrus.WarnLevel)}, false},
		{"Fatal", args{Level(logrus.FatalLevel)}, false},
		{"nil string", args{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotValid := isLogLevelValid(tt.args.logLevel); gotValid != tt.wantValid {
				t.Errorf("isLogLevelValid() = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}

func emptyBufferBackend() io.Writer {
	return bytes.NewBufferString("")
}
func Test_newLogger_logLevel(t *testing.T) {
	type args struct {
		logLevel    logrus.Level
		backend     io.Writer
		packageName string
	}
	tests := []struct {
		name         string
		args         args
		wantLogLevel Level
		wantErr      bool
	}{
		{"Debug", args{logrus.DebugLevel, StdoutBackend, "test"}, DebugLevel, false},
		{"Info", args{logrus.InfoLevel, StdoutBackend, "test"}, InfoLevel, false},
		{"Error", args{logrus.ErrorLevel, StdoutBackend, "test"}, ErrorLevel, false},

		{"Debug", args{logrus.DebugLevel, StderrBackend, "test"}, DebugLevel, false},
		{"Info", args{logrus.InfoLevel, StderrBackend, "test"}, InfoLevel, false},
		{"Error", args{logrus.ErrorLevel, StderrBackend, "test"}, ErrorLevel, false},

		{"Debug", args{logrus.DebugLevel, emptyBufferBackend(), "test"}, DebugLevel, false},
		{"Info", args{logrus.InfoLevel, emptyBufferBackend(), "test"}, InfoLevel, false},
		{"Error", args{logrus.ErrorLevel, emptyBufferBackend(), "test"}, ErrorLevel, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger, err := newLogger(tt.args.logLevel, tt.args.backend, tt.args.packageName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("newLogger() error = %v, wantErr %v", err, tt.wantErr)
			}

			if (err != nil) && tt.wantErr {
				t.Logf("newLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotLogLevel := logger.GetLogLevel()
			if gotLogLevel != tt.wantLogLevel {
				t.Errorf("newLogger() gotLogLevel = %v, wantLogLevel %v", gotLogLevel, tt.wantLogLevel)
			}

		})
	}
}

func Test_NewLogger_logLevel(t *testing.T) {
	type args struct {
		logLevel    Level
		backend     io.Writer
		packageName string
	}
	tests := []struct {
		name         string
		args         args
		wantLogLevel Level
		wantErr      bool
	}{
		{"Debug", args{DebugLevel, StdoutBackend, "test"}, DebugLevel, false},
		{"Info", args{InfoLevel, StdoutBackend, "test"}, InfoLevel, false},
		{"Error", args{ErrorLevel, StdoutBackend, "test"}, ErrorLevel, false},

		{"Debug", args{DebugLevel, StderrBackend, "test"}, DebugLevel, false},
		{"Info", args{InfoLevel, StderrBackend, "test"}, InfoLevel, false},
		{"Error", args{ErrorLevel, StderrBackend, "test"}, ErrorLevel, false},

		{"Debug", args{DebugLevel, emptyBufferBackend(), "test"}, DebugLevel, false},
		{"Info", args{InfoLevel, emptyBufferBackend(), "test"}, InfoLevel, false},
		{"Error", args{ErrorLevel, emptyBufferBackend(), "test"}, ErrorLevel, false},

		{"Debug", args{Level(logrus.WarnLevel), StderrBackend, "test"}, DebugLevel, true},
		{"Debug", args{Level(logrus.FatalLevel), StderrBackend, "test"}, DebugLevel, true},

		{"Debug", args{Level(logrus.WarnLevel), StdoutBackend, "test"}, DebugLevel, true},
		{"Debug", args{Level(logrus.FatalLevel), StdoutBackend, "test"}, DebugLevel, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger, err := NewLogger(tt.args.logLevel, tt.args.backend, tt.args.packageName)

			if (err != nil) != tt.wantErr {
				t.Fatalf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
			}

			if (err != nil) && tt.wantErr {
				t.Logf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if logger == nil {
				t.Fatalf("NewLogger() error = nil, but logger = nil, want not nil")
			}

			gotLogLevel := logger.GetLogLevel()
			if gotLogLevel != tt.wantLogLevel {
				t.Errorf("NewLogger() gotLogLeveL = %v, wantLogLevel %v", gotLogLevel, tt.wantLogLevel)
			}

		})
	}
}
func Test_Logger_Methods(t *testing.T) {
	type setupArgs struct {
		logLevel    logrus.Level
		backend     *bytes.Buffer
		packageName string
	}
	tests := []struct {
		name           string
		setupArgs      setupArgs
		logFunc        func(LoggerInterface, ...interface{})
		wantDataLogged bool
	}{
		{name: "Debug_DebugLevel",
			setupArgs:      setupArgs{logrus.DebugLevel, &bytes.Buffer{}, "test"},
			logFunc:        LoggerInterface.Debug,
			wantDataLogged: true},
		{name: "Debug_InfoLevel",
			setupArgs:      setupArgs{logrus.InfoLevel, &bytes.Buffer{}, "test"},
			logFunc:        LoggerInterface.Debug,
			wantDataLogged: false},
		{name: "Debug_ErrorLevel",
			setupArgs:      setupArgs{logrus.ErrorLevel, &bytes.Buffer{}, "test"},
			logFunc:        LoggerInterface.Debug,
			wantDataLogged: false},

		{name: "Info_DebugLevel",
			setupArgs:      setupArgs{logrus.DebugLevel, &bytes.Buffer{}, "test"},
			logFunc:        LoggerInterface.Info,
			wantDataLogged: true},
		{name: "Info_InfoLevel",
			setupArgs:      setupArgs{logrus.InfoLevel, &bytes.Buffer{}, "test"},
			logFunc:        LoggerInterface.Info,
			wantDataLogged: true},
		{name: "Info_ErrorLevel",
			setupArgs:      setupArgs{logrus.ErrorLevel, &bytes.Buffer{}, "test"},
			logFunc:        LoggerInterface.Info,
			wantDataLogged: false},

		{name: "Error_DebugLevel",
			setupArgs:      setupArgs{logrus.DebugLevel, &bytes.Buffer{}, "test"},
			logFunc:        LoggerInterface.Error,
			wantDataLogged: true},
		{name: "Error_InfoLevel",
			setupArgs:      setupArgs{logrus.InfoLevel, &bytes.Buffer{}, "test"},
			logFunc:        LoggerInterface.Error,
			wantDataLogged: true},
		{name: "Error_ErrorLevel",
			setupArgs:      setupArgs{logrus.ErrorLevel, &bytes.Buffer{}, "test"},
			logFunc:        LoggerInterface.Error,
			wantDataLogged: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			logger, err := newLogger(tt.setupArgs.logLevel, tt.setupArgs.backend, tt.setupArgs.packageName)
			if err != nil {
				t.Fatalf("logger setup error - %v", err)
			}

			tt.logFunc(logger, "test-data")

			gotLog := strings.Split(tt.setupArgs.backend.String(), "\n")
			gotLogData := false
			if 2 == len(gotLog) { //Expected log entry + new line
				gotLogData = true
			}

			if gotLogData != tt.wantDataLogged {
				t.Errorf("Debug() Expected log at %v, got nil", tt.setupArgs.logLevel)
			}

		})
	}
}
