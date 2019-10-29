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
	"testing"
	"time"

	"github.com/spf13/pflag"

	"github.com/direct-state-transfer/dst-go/ethereum/types"
)

func setupNewFlag(flagTypes ...string) *pflag.FlagSet {
	var fs = pflag.NewFlagSet("test", pflag.ContinueOnError)
	for _, flagType := range flagTypes {

		switch flagType {
		case "bool":
			fs.Bool("boolFlag", false, "")
		case "boolSlice":
			fs.BoolSlice("boolSliceFlag", []bool{}, "")
		case "bytesHex":
			fs.BytesHex("bytesHexFlag", []byte{}, "")
		case "float32":
			fs.Float32("float32Flag", 0, "")
		case "float64":
			fs.Float64("float64Flag", 0, "")
		case "int8":
			fs.Int8("int8Flag", 0, "")
		case "int16":
			fs.Int16("int16Flag", 0, "")
		case "int32":
			fs.Int32("int32Flag", 0, "")
		case "int64":
			fs.Int64("int64Flag", 0, "")
		case "int":
			fs.Int("intFlag", 0, "")
		case "intSlice":
			fs.IntSlice("intSliceFlag", []int{}, "")
		case "netIP":
			fs.IP("netIPFlag", net.IP{}, "")
		case "netIPMask":
			fs.IPMask("netIPMaskFlag", net.IPMask{}, "")
		case "netIPNet":
			fs.IPNet("netIPNetFlag", net.IPNet{}, "")
		case "netIPSlice":
			fs.IPSlice("netIPSliceFlag", []net.IP{}, "")
		case "uint8":
			fs.Uint8("uint8Flag", 0, "")
		case "uint16":
			fs.Uint16("uint16Flag", 0, "")
		case "uint32":
			fs.Uint32("uint32Flag", 0, "")
		case "uint64":
			fs.Uint64("uint64Flag", 0, "")
		case "uint":
			fs.Uint("uintFlag", 0, "")
		case "uintSlice":
			fs.UintSlice("uintSliceFlag", []uint{}, "")
		case "string":
			fs.String("stringFlag", "", "")
		case "stringArray":
			fs.StringArray("stringArrayFlag", []string{}, "")
		case "timeDuration":
			fs.Duration("timeDurationFlag", time.Duration(0), "")
		case "timeDurationSlice":
			fs.DurationSlice("timeDurationSliceFlag", []time.Duration{}, "")
		default:
		}
	}
	return fs
}

func setupParseFlags(fs *pflag.FlagSet, argsToParse []string) (err error) {
	for _, arg := range argsToParse {

		stringToParse := fmt.Sprint(arg)
		err = fs.Parse([]string{stringToParse})
		if err != nil {
			break
		}
	}
	return err
}

type unsupportedTypeForTest int

func Test_Lookup(t *testing.T) {

	type flagInfo struct {
		name      string
		targetVar interface{}
	}

	type args struct {
		flagSet  *pflag.FlagSet
		flagInfo []flagInfo
	}
	tests := []struct {
		name          string
		args          args
		stringToParse []string
		wantChanged   bool
		wantErr       bool
	}{
		//Positive test cases
		{
			name: "boolFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("bool"),
				flagInfo: []flagInfo{
					{"boolFlag", new(bool)},
				},
			},
			stringToParse: []string{"--boolFlag=true"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "boolFlag_ValueUnspecified_FlagDefined",
			args: args{
				flagSet: setupNewFlag("bool"),
				flagInfo: []flagInfo{
					{"boolFlag", new(bool)},
				},
			},
			stringToParse: []string{""},
			wantChanged:   false,
			wantErr:       false,
		},
		{
			name: "boolSliceFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("boolSlice"),
				flagInfo: []flagInfo{
					{"boolSliceFlag", new([]bool)},
				},
			},
			stringToParse: []string{fmt.Sprintf("--boolSliceFlag=%s",
				strings.Join([]string{"1", "F", "TRUE", "0"}, ","))},
			wantChanged: true,
			wantErr:     false,
		},
		{
			name: "boolSliceFlag_ValueUnspecified_FlagtDefined",
			args: args{
				flagSet: setupNewFlag("boolSlice"),
				flagInfo: []flagInfo{
					{"boolSliceFlag", new([]bool)},
				},
			},
			stringToParse: []string{fmt.Sprintf("--boolSliceFlag=%s",
				strings.Join([]string{"1", "F", "TRUE", "0"}, ","))},
			wantChanged: true,
			wantErr:     false,
		},
		{
			name: "bytesHexFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("bytesHex"),
				flagInfo: []flagInfo{
					{"bytesHexFlag", new([]byte)},
				},
			},
			stringToParse: []string{"--bytesHexFlag=2345ABCD"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "float32Flag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("float32"),
				flagInfo: []flagInfo{
					{"float32Flag", new(float32)},
				},
			},
			stringToParse: []string{"--float32Flag=32.50"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "float64Flag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("float64"),
				flagInfo: []flagInfo{
					{"float64Flag", new(float64)},
				},
			},
			stringToParse: []string{"--float64Flag=63.43"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "int8Flag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("int8"),
				flagInfo: []flagInfo{
					{"int8Flag", new(int8)},
				},
			},
			stringToParse: []string{"--int8Flag=8"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "int16Flag_ValueSpecified",
			args: args{

				flagSet: setupNewFlag("int16"),
				flagInfo: []flagInfo{
					{"int16Flag", new(int16)},
				},
			},
			stringToParse: []string{"--int16Flag=16"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "int32Flag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("int32"),
				flagInfo: []flagInfo{
					{"int32Flag", new(int32)},
				},
			},
			stringToParse: []string{"--int32Flag=32"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "int64Flag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("int64"),
				flagInfo: []flagInfo{
					{"int64Flag", new(int64)},
				},
			},
			stringToParse: []string{"--int64Flag=64"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "intFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("int"),
				flagInfo: []flagInfo{
					{"intFlag", new(int)},
				},
			},
			stringToParse: []string{"--intFlag=1"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "intSliceFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("intSlice"),
				flagInfo: []flagInfo{
					{"intSliceFlag", new([]int)},
				},
			},
			stringToParse: []string{fmt.Sprintf("--intSliceFlag=%s",
				strings.Join([]string{"1", "2", "3", "4"}, ","))},
			wantChanged: true,
			wantErr:     false,
		},
		{
			name: "netIPFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("netIP"),
				flagInfo: []flagInfo{
					{"netIPFlag", new(net.IP)},
				},
			},
			stringToParse: []string{"--netIPFlag=192.168.1.1"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "netIPMaskFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("netIPMask"),
				flagInfo: []flagInfo{
					{"netIPMaskFlag", new(net.IPMask)},
				},
			},
			stringToParse: []string{"--netIPMaskFlag=192.168.1.1"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "netIPNetFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("netIPNet"),
				flagInfo: []flagInfo{
					{"netIPNetFlag", new(net.IPNet)},
				},
			},
			stringToParse: []string{"--netIPNetFlag=192.168.1.1/16"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "netIPSliceFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("netIPSlice"),
				flagInfo: []flagInfo{
					{"netIPSliceFlag", new([]net.IP)},
				},
			},
			stringToParse: []string{fmt.Sprintf("--netIPSliceFlag=%s",
				strings.Join([]string{"127.0.0.1", "0.0.0.0",
					"a4ab:61d:f03e:5d7d:fad7:d4c2:a1a5:568"}, ","))},
			wantChanged: true,
			wantErr:     false,
		},
		{
			name: "stringFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("string"),
				flagInfo: []flagInfo{
					{"stringFlag", new(string)},
				},
			},
			stringToParse: []string{"--stringFlag=testArg"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "stringArrayFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("stringArray"),
				flagInfo: []flagInfo{
					{"stringArrayFlag", new([]string)},
				},
			},
			stringToParse: []string{fmt.Sprintf("--stringArrayFlag=%s",
				strings.Join([]string{"1", "2", "3", "4"}, ","))},
			wantChanged: true,
			wantErr:     false,
		},
		{
			name: "timeDurationFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("timeDuration"),
				flagInfo: []flagInfo{
					{"timeDurationFlag", new(time.Duration)},
				},
			},
			stringToParse: []string{"--timeDurationFlag=3m"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "timeDurationSliceFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("timeDurationSlice"),
				flagInfo: []flagInfo{
					{"timeDurationSliceFlag", new([]time.Duration)},
				},
			},
			stringToParse: []string{fmt.Sprintf("--timeDurationSliceFlag=%s",
				strings.Join([]string{"1h", "2m", "3s", "4ns"}, ","))},
			wantChanged: true,
			wantErr:     false,
		},
		{
			name: "uint8Flag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("uint8"),
				flagInfo: []flagInfo{
					{"uint8Flag", new(uint8)},
				},
			},
			stringToParse: []string{"--uint8Flag=8"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "uint16Flag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("uint16"),
				flagInfo: []flagInfo{
					{"uint16Flag", new(uint16)},
				},
			},
			stringToParse: []string{"--uint16Flag=16"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "uint32Flag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("uint32"),
				flagInfo: []flagInfo{
					{"uint32Flag", new(uint32)},
				},
			},
			stringToParse: []string{"--uint32Flag=32"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "uint64Flag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("uint64"),
				flagInfo: []flagInfo{
					{"uint64Flag", new(uint64)},
				},
			},
			stringToParse: []string{"--uint64Flag=64"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "uintFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("uint"),
				flagInfo: []flagInfo{
					{"uintFlag", new(uint)},
				},
			},
			stringToParse: []string{"--uintFlag=1"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "uintSliceFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("uintSlice"),
				flagInfo: []flagInfo{
					{"uintSliceFlag", new([]uint)},
				},
			},
			stringToParse: []string{fmt.Sprintf("--uintSliceFlag=%s",
				strings.Join([]string{"1", "2", "3", "4"}, ","))},
			wantChanged: true,
			wantErr:     false,
		},

		//Negative test cases, passes when error is returned
		{
			name: "commonAddressFlag_ValueSpecified",
			args: args{
				flagSet: setupNewFlag("string"),
				flagInfo: []flagInfo{
					{"stringFlag", new(types.Address)},
				},
			},
			stringToParse: []string{"--stringFlag=0x6FEe9f34c2b9fb85539Be5747e9001A736884793"},
			wantChanged:   true,
			wantErr:       false,
		},
		{
			name: "unsupported_type_value_specified",
			args: args{
				flagSet: setupNewFlag("string"),
				flagInfo: []flagInfo{
					{"stringFlag", new(unsupportedTypeForTest)},
				},
			},
			stringToParse: []string{"--stringFlag=0x6FEe9f34c2b9fb85539Be5747e9001A736884793"},
			wantChanged:   false,
			wantErr:       true,
		},
		{
			name: "unsupported_type_value_unspecified",
			args: args{
				flagSet: setupNewFlag("string"),
				flagInfo: []flagInfo{
					{"stringFlag", new(unsupportedTypeForTest)},
				},
			},
			stringToParse: []string{"--stringFlag=0x6FEe9f34c2b9fb85539Be5747e9001A736884793"},
			wantChanged:   false,
			wantErr:       true,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			err := setupParseFlags(tt.args.flagSet, tt.stringToParse)
			if err != nil {
				t.Fatalf("Setup error -  %v", err)
			}

			for _, flag := range tt.args.flagInfo {
				gotChanged, err := Lookup(tt.args.flagSet, flag.name, flag.targetVar)
				if (err != nil) != tt.wantErr {
					t.Fatalf("Lookup() error = %v, wantErr %v", err, tt.wantErr)
				}

				if (err != nil) && tt.wantErr {
					t.Logf("Lookup() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if gotChanged != tt.wantChanged {
					t.Errorf("Lookup() = %v, want %v", gotChanged, tt.wantChanged)
				}
			}
		})
	}
}

func Test_LookupMultiple(t *testing.T) {

	type args struct {
		flagSet   *pflag.FlagSet
		flagsInfo []FlagInfo
	}
	tests := []struct {
		name          string
		args          args
		stringToParse []string
		wantChanged   []bool
		wantErr       bool
	}{
		//Positive test cases
		{
			name: "multipleFlags_allValuesSpecified_flagsDefined",
			args: args{
				flagSet: setupNewFlag("uint", "netIP", "bool"),
				flagsInfo: []FlagInfo{
					{"uintFlag", new(uint), false},
					{"netIPFlag", new(net.IP), false},
					{"boolFlag", new(bool), false},
				},
			},
			stringToParse: []string{"--uintFlag=101", "--netIPFlag=127.0.0.1", "--boolFlag=TRUE"},
			wantChanged:   []bool{true, true, true},
			wantErr:       false,
		},
		{
			name: "multipleFlags_fewValuesSpecified_flagsDefined",
			args: args{
				flagSet: setupNewFlag("uint", "netIP", "bool"),
				flagsInfo: []FlagInfo{
					{"uintFlag", new(uint), false},
					{"netIPFlag", new(net.IP), true},
					{"boolFlag", new(bool), false},
				},
			},
			stringToParse: []string{"--uintFlag=101", "--boolFlag=TRUE"},
			wantChanged:   []bool{true, false, true},
			wantErr:       false,
		}, {
			name: "multipleFlags_unsupportedFlagDefined",
			args: args{
				flagSet: setupNewFlag("string"),
				flagsInfo: []FlagInfo{
					{"stringFlag", new(unsupportedTypeForTest), false},
				},
			},
			stringToParse: []string{"--stringFlag=0x6FEe9f34c2b9fb85539Be5747e9001A736884793"},
			wantChanged:   []bool{false},
			wantErr:       true,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			err := setupParseFlags(tt.args.flagSet, tt.stringToParse)
			if err != nil {
				t.Fatalf("Setup error -  %s", err)
			}

			err = LookUpMultiple(tt.args.flagSet, tt.args.flagsInfo)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LookupMultiple() error = %v, wantErr %v", err, tt.wantErr)
			}

			if (err != nil) && tt.wantErr {
				t.Logf("LookupMultiple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for idx, flagInfo := range tt.args.flagsInfo {
				if flagInfo.Changed != tt.wantChanged[idx] {
					t.Errorf("LookupMultiple() flag %s changed = %v, wantChanged %v",
						flagInfo.Name, flagInfo.Changed, tt.wantChanged[idx])
				}
			}
		})
	}
}
