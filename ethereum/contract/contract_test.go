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

package contract

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func fsSetupWithContractsAll(contractList []Handler, testContractsDir string) (fs afero.Fs, err error) {
	_ = contractList
	fs = afero.NewMemMapFs()

	_ = fs.Mkdir(testContractsDir, os.ModeDir)
	for index := range contractList {

		contract := &contractList[index]
		randomByteArray := make([]byte, 250)
		createFile := func(hashExpected *string, fileToCreate fileType) {
			_, _ = rand.Read(randomByteArray)
			*hashExpected = fmt.Sprintf("%x", sha256.Sum256(randomByteArray))
			absFilePath, _ := contract.getAbsFilepath(testContractsDir, fileToCreate)
			_ = afero.WriteFile(fs, absFilePath, randomByteArray, 0644)
		}

		createFile(&contract.HashGoFile, golangFile)
		createFile(&contract.HashBinRuntimeFile, binRuntimeFile)
		createFile(&contract.HashSolFile, solidityFile)
	}

	return fs, nil
}

func fsSetupEmpty(contractList []Handler, testContractsDir string) (afero.Fs, error) {
	_ = contractList
	return afero.NewMemMapFs(), nil
}

func fsSetupCorruptHash(contractList []Handler, testContractsDir string) (afero.Fs, error) {
	fs, err := fsSetupWithContractsAll(contractList, testContractsDir)
	if err != nil {
		return nil, err
	}
	for index := range contractList {

		contract := &contractList[index]
		contract.HashSolFile = contract.HashGoFile
		contract.HashGoFile = contract.HashBinRuntimeFile
	}
	return fs, err
}

func Test_ValidateIntegrity(t *testing.T) {
	type args struct {
		fs           afero.Fs
		contractList []Handler
	}
	tests := []struct {
		name             string
		args             args
		setupFunc        func([]Handler, string) (afero.Fs, error)
		wantCorruptFiles bool
		wantMissingFiles bool
		wantErr          bool
	}{
		{
			name: "all_valid",
			args: args{contractList: []Handler{{
				Name:               "LibSignatures",
				HashSolFile:        strings.ToLower("359e2e9f7bacdcefc6962c46182aba7f16b8b0a8314468ca8dd88edd25299209"),
				HashGoFile:         strings.ToLower("1b6a7102bd6726168ea973b0bb2107655b638bc09b702cffe05a9d90e67373d7"),
				HashBinRuntimeFile: strings.ToLower("3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9"),
				GasUnits:           uint64(40e5),
				Version:            "0.0.1",
			}}},
			setupFunc:        fsSetupWithContractsAll,
			wantCorruptFiles: false,
			wantMissingFiles: false,
			wantErr:          false,
		},
		{
			name: "empty",
			args: args{contractList: []Handler{{
				Name:               "LibSignatures",
				HashSolFile:        strings.ToLower("359e2e9f7bacdcefc6962c46182aba7f16b8b0a8314468ca8dd88edd25299209"),
				HashGoFile:         strings.ToLower("1b6a7102bd6726168ea973b0bb2107655b638bc09b702cffe05a9d90e67373d7"),
				HashBinRuntimeFile: strings.ToLower("3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9"),
				GasUnits:           uint64(40e5),
				Version:            "0.0.1",
			}}},
			setupFunc:        fsSetupEmpty,
			wantCorruptFiles: false,
			wantMissingFiles: true,
			wantErr:          true,
		},
		{
			name: "corrupt_files",
			args: args{contractList: []Handler{{
				Name:               "LibSignatures",
				HashSolFile:        strings.ToLower("359e2e9f7bacdcefc6962c46182aba7f16b8b0a8314468ca8dd88edd25299209"),
				HashGoFile:         strings.ToLower("1b6a7102bd6726168ea973b0bb2107655b638bc09b702cffe05a9d90e67373d7"),
				HashBinRuntimeFile: strings.ToLower("3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9"),
				GasUnits:           uint64(40e5),
				Version:            "0.0.1",
			}}},
			setupFunc:        fsSetupCorruptHash,
			wantCorruptFiles: true,
			wantMissingFiles: false,
			wantErr:          true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			//Setup
			fs, err := tt.setupFunc(tt.args.contractList, testContractsDir)
			if err != nil {
				t.Fatalf("validateIntegrity() could not setup test file system. error : %v", err)
			}
			tt.args.fs = fs

			gotCorruptFiles, gotMissingFiles, err := ValidateIntegrity(tt.args.fs, tt.args.contractList, testContractsDir)

			if (err != nil) != tt.wantErr {
				t.Fatalf("validateIntegrity() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (gotCorruptFiles != nil) != tt.wantCorruptFiles {
				t.Errorf("validateIntegrity() gotCorruptFiles = %v, want %v", gotCorruptFiles, tt.wantCorruptFiles)
			}
			if (gotMissingFiles != nil) != tt.wantMissingFiles {
				t.Errorf("validateIntegrity() gotMissingFiles = %v, want %v", gotMissingFiles, tt.wantMissingFiles)
			}
		})
	}
}

func Test_CheckSha256Sum(t *testing.T) {
	type args struct {
		reader       io.Reader
		reqSha256Sum string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus MatchStatus
		wantErr    bool
	}{
		{
			"match1-ShaInUpperCase",
			args{strings.NewReader("This is a test string 123!@#"), "35E624638722DB15BAE66E1E33CC746FDB58C48B349FB14DB6977EF30B4F613A"},
			Match, false,
		},
		{
			"match2-ShaInLowerCase",
			args{strings.NewReader("This is a test string 123!@#"), "35e624638722db15bae66e1e33cc746fdb58c48b349fb14db6977ef30b4f613a"},
			Match, false,
		},
		{
			"match3-ShaInMixedCase",
			args{strings.NewReader("This is a test string 123!@#"), "35e624638722DB15BAE66e1e33cc746FDB58C48B349FB14DB6977EF30B4F613A"},
			Match, false,
		},
		{
			"no-match-validSha",
			args{strings.NewReader("This is a test string 123!@#"), "22DB15BAE66E135E62448B349FB0B4F613A6387E33CC746F14DB6977EF3DB58C"},
			NoMatch, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, err := CheckSha256Sum(tt.args.reader, tt.args.reqSha256Sum)

			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckSha256Sum() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("CheckSha256Sum() = %v, want %v", gotStatus, tt.wantStatus)
			}
		})
	}
}
