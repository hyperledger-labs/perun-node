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
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Enumeration of supported log backends
var (
	StdoutBackend = os.Stdout
	StderrBackend = os.Stderr
)

// Level represents the log level type
type Level logrus.Level

// String implements fmt.Stringer interface
func (level Level) String() string {
	return logrus.Level(level).String()
}

// Enumeration of supported log levels
var (
	DebugLevel = Level(logrus.DebugLevel)
	InfoLevel  = Level(logrus.InfoLevel)
	ErrorLevel = Level(logrus.ErrorLevel)

	validLevels = []Level{DebugLevel, InfoLevel, ErrorLevel}
)

// Config represents the configuration parameters for the logger
type Config struct {
	Level         Level
	Backend       io.Writer
	LevelString   string
	BackendString string
}

// GetLogLevel returns the log level from logger configuration
func (cfg Config) GetLogLevel() Level {
	return cfg.Level
}

// SetLogLevel sets the log level in logger configuration
func (cfg *Config) SetLogLevel(newLevel Level) (err error) {
	if !isLogLevelValid(newLevel) {
		return fmt.Errorf("invalid log level - %s", newLevel)
	}
	cfg.Level = newLevel
	return nil
}

// Configurer represents the methods to read/write logger configuration
type Configurer interface {
	GetLogLevel() Level
	SetLogLevel(Level) (err error)
}

// LoggerInterface represents methods that can be used by other packages to log messages.
// These methods are independent of the actual logging library used
type LoggerInterface interface {
	GetLogLevel() Level

	Debug(...interface{})
	Info(...interface{})
	Error(...interface{})
}

// Logger represents the logger type that wraps the actual logger form the logging library.
// It also adds additional fields to it.
type Logger struct {
	*logrus.Logger // Logger represents the actual instance of logger from logging library used.

	packageName string // package name will be added to each log entry.
}

// GetLogLevel returns the log level of the logger.
func (l *Logger) GetLogLevel() (level Level) {
	if l != nil && l.Logger != nil {
		level = Level(l.Level)
	}
	return
}

// Debug logs the message at debug level.
// The message is formatted with default format for its operands.
func (l *Logger) Debug(args ...interface{}) {
	if l != nil && l.Logger != nil {
		l.WithFields(logrus.Fields{"package": l.packageName}).Debug(args...)
	}
}

// Info logs the message at info level.
// The message is formatted with default format for its operands.
func (l *Logger) Info(args ...interface{}) {
	if l != nil && l.Logger != nil {
		l.WithFields(logrus.Fields{"package": l.packageName}).Info(args...)
	}
}

// Error logs the message at error level.
// The message is formatted with default format for its operands.
func (l *Logger) Error(args ...interface{}) {
	if l != nil && l.Logger != nil {
		l.WithFields(logrus.Fields{"package": l.packageName}).Error(args...)
	}
}

// NewLogger returns an initialized logger instance configured with provided values.
func NewLogger(logLevelStr Level, backendStr io.Writer, packageName string) (logger LoggerInterface, err error) {

	if !isLogLevelValid(logLevelStr) {
		return nil, fmt.Errorf("invalid log level - %s", logLevelStr)
	}

	return newLogger(logrus.Level(logLevelStr), backendStr, packageName)

}

func isLogLevelValid(logLevel Level) bool {
	for _, level := range validLevels {
		if level == logLevel {
			return true
		}
	}
	return false
}

func newLogger(logLevel logrus.Level, backend io.Writer, packageName string) (logger LoggerInterface, err error) {

	newLogger := &Logger{
		Logger: logrus.New()}

	newLogger.SetFormatter(&customTextFormatter{logrus.TextFormatter{
		FullTimestamp:          true,
		TimestampFormat:        "2006-01-02 15:04:05 Z0700",
		DisableLevelTruncation: true,
	}})

	newLogger.SetLevel(logLevel)

	newLogger.SetOutput(backend)

	newLogger.packageName = packageName

	return newLogger, nil
}

// customTextFormatter is defined to override default formating options for log entry.
type customTextFormatter struct {
	logrus.TextFormatter
}

// Format modifies the default logging format.
func (f *customTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	originalText, err := f.TextFormatter.Format(entry)
	return append([]byte("▶ "), originalText...), err
}
