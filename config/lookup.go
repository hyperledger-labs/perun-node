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

package config

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/direct-state-transfer/dst-go/ethereum/types"
)

// FlagInfo represents flag information consisting of flag name, pointer to store its value and a variable to track if the flag was modified.
type FlagInfo struct {
	Name    string
	Ptr     interface{}
	Changed bool
}

// LookUpMultiple parses the values of flag defined in each of the targets.
func LookUpMultiple(flagSet *pflag.FlagSet, targets []FlagInfo) (err error) {

	for idx := range targets {
		target := &targets[idx]
		target.Changed, err = Lookup(flagSet, target.Name, target.Ptr)
		if err != nil {
			err = fmt.Errorf("lookup flag %s error - %s", target.Name, err)
		}
	}
	return err
}

// Lookup parses the values of flag to targetVar if it defined in flagSet and its values has been modified.
// Status of whether the flag was modified or not is returned in the changed.
func Lookup(flagSet *pflag.FlagSet, name string, targetVar interface{}) (
	changed bool, err error) {

	if changed = flagSet.Changed(name); changed {

		switch t := targetVar.(type) {
		case *bool:
			*t, err = flagSet.GetBool(name)
		case *[]bool:
			*t, err = flagSet.GetBoolSlice(name)
		case *[]byte:
			//bytes flag type supports hex encoded string
			*t, err = flagSet.GetBytesHex(name)
		case *float32:
			*t, err = flagSet.GetFloat32(name)
		case *float64:
			*t, err = flagSet.GetFloat64(name)
		case *int8:
			*t, err = flagSet.GetInt8(name)
		case *int16:
			*t, err = flagSet.GetInt16(name)
		case *int32:
			*t, err = flagSet.GetInt32(name)
		case *int64:
			*t, err = flagSet.GetInt64(name)
		case *int:
			*t, err = flagSet.GetInt(name)
		case *[]int:
			*t, err = flagSet.GetIntSlice(name)
		case *net.IP:
			*t, err = flagSet.GetIP(name)
			//TODO : Figure out difference between Ip and IpMask
		case *net.IPMask:
			*t, err = flagSet.GetIPv4Mask(name)
		case *net.IPNet:
			*t, err = flagSet.GetIPNet(name)
		case *[]net.IP:
			*t, err = flagSet.GetIPSlice(name)
		case *string:
			*t, err = flagSet.GetString(name)
		case *[]string:
			*t, err = flagSet.GetStringArray(name)
		case *time.Duration:
			*t, err = flagSet.GetDuration(name)
		case *[]time.Duration:
			*t, err = flagSet.GetDurationSlice(name)
		case *uint8:
			*t, err = flagSet.GetUint8(name)
		case *uint16:
			*t, err = flagSet.GetUint16(name)
		case *uint32:
			*t, err = flagSet.GetUint32(name)
		case *uint64:
			*t, err = flagSet.GetUint64(name)
		case *uint:
			*t, err = flagSet.GetUint(name)
		case *[]uint:
			*t, err = flagSet.GetUintSlice(name)
		case *types.Address:
			//parse the values as string and convert it to address
			var addrStr string
			addrStr, err = flagSet.GetString(name)
			*t = types.HexToAddress(strings.ToLower(addrStr))
		default:
			err = fmt.Errorf("unsupported data type (%T) for flag", t)
			changed = false
		}
	}
	return changed, err
}
