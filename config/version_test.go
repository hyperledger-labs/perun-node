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

import "testing"

var (
	testVersion = Version{
		Major: 0,
		Minor: 1,
		Patch: 0,
		Meta:  "unstable",
	}
	testVersionString = "0.1.0-unstable"
)

func Test_Version_String(t *testing.T) {
	tests := []struct {
		name string
		v    Version
		want string
	}{
		{
			name: "versionString",
			v:    testVersion,
			want: testVersionString,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Version_StringWithCommitID(t *testing.T) {
	type args struct {
		commitID string
	}
	tests := []struct {
		name string
		v    Version
		args args
		want string
	}{
		{
			name: "emptyString",
			v:    testVersion,
			args: args{
				commitID: "",
			},
			want: testVersionString + "",
		},
		{
			name: "lessThan8Char",
			v:    testVersion,
			args: args{
				commitID: "1234567",
			},
			want: testVersionString + "-1234567",
		},
		{
			name: "equalTo8Char",
			v:    testVersion,
			args: args{
				commitID: "12345678",
			},
			want: testVersionString + "-12345678",
		},
		{
			name: "greaterThan8Char",
			v:    testVersion,
			args: args{
				commitID: "ABCDEF123456789",
			},
			want: testVersionString + "-ABCDEF12",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.StringWithCommitID(tt.args.commitID); got != tt.want {
				t.Errorf("Version.StringWithCommitID() = %v, want %v", got, tt.want)
			}
		})
	}
}
