// Copyright (c) 2019 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/dst-go
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

package adapter

import (
	"github.com/spf13/pflag"

	"github.com/direct-state-transfer/dst-go/config"
	"github.com/direct-state-transfer/dst-go/log"
)

// Config represents the configuration for this module.
type Config struct {
	Logger  log.Config
	maxConn uint64
}

// ConfigDefault represents the default configuration for this module.
var ConfigDefault = Config{
	maxConn: 100,
}

// GetFlagSet initializes and returns a flagset with flags for configuring this module.
func GetFlagSet() *pflag.FlagSet {

	var chFlags pflag.FlagSet

	chFlags.Uint64(
		"maxChConn", 0, "Maximum number of channel connections allowed")
	chFlags.String(
		"channelLogLevel", "", "Log level for channel module")
	chFlags.String(
		"channelLogBackend", "", "Log Backend for channel module")
	return &chFlags
}

// ParseFlags parses the flags defined in this module.
func ParseFlags(flagSet *pflag.FlagSet, cfg *Config) (err error) {

	var flagsToParse = []config.FlagInfo{
		{Name: "channelLogLevel", Ptr: &cfg.Logger.Level},
		{Name: "channelLogBackend", Ptr: &cfg.Logger.Backend},
		{Name: "maxChConn", Ptr: &cfg.maxConn},
	}

	return config.LookUpMultiple(flagSet, flagsToParse)
}
