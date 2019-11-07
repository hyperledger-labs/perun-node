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
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// Store represents the instance of contract store.
var Store = StoreType{
	contractsDir: "contract_store",
	libSignatures: Handler{
		Name:               "LibSignatures",
		HashSolFile:        strings.ToLower("359e2e9f7bacdcefc6962c46182aba7f16b8b0a8314468ca8dd88edd25299209"),
		HashGoFile:         strings.ToLower("1b6a7102bd6726168ea973b0bb2107655b638bc09b702cffe05a9d90e67373d7"),
		HashBinRuntimeFile: strings.ToLower("3c0f29dfe76fd55ab0b023b26c97d2a306805a03788277cd0c7d2817cb7a9bf9"),
		GasUnits:           uint64(40e5),
		Version:            "0.0.1",
	},
	msContract: Handler{
		Name:               "MSContract",
		HashSolFile:        strings.ToLower("d2f7c0055a445f4823a2a4312df1bb7602fb62b022d99047634fe0cb0a941938"),
		HashGoFile:         strings.ToLower("8a7f2d750db8bb1bf7c1a8fedae204a828a1e81e7ac7a252737f3528dea29ceb"),
		HashBinRuntimeFile: strings.ToLower("4fb304c42b1bad1b03417c72e925d87488dea1d22e13ed0601abbe8d86c8e8ad"),
		GasUnits:           uint64(40e5),
		Version:            "0.0.1",
	},
	vpc: Handler{
		Name:               "VPC",
		HashSolFile:        strings.ToLower("c2195856c9d206c18d3ec49eb8b4b38a2b022ec6ef155478fe1bb3e8fd78b494"),
		HashGoFile:         strings.ToLower("1b6a7102bd6726168ea973b0bb2107655b638bc09b702cffe05a9d90e67373d7"),
		HashBinRuntimeFile: strings.ToLower("978bc824e314529d620afb0c1f07770dad5e36b11ad0060102f67fc29e3a60fe"),
		GasUnits:           uint64(40e5),
		Version:            "0.0.1",
	},
	timeoutMSContract:          100 * time.Minute,
	timeoutVPCValidity:         10 * time.Minute,
	timeoutVPCExtendedValidity: 20 * time.Minute,
}

// CheckSha256Sum returns Match if SHA256 sum over the data in reader matches the reqSHA256 sum.
// Else a NoMatch or Unknown status with appropriate error information is returned.
func CheckSha256Sum(reader io.Reader, reqSha256Sum string) (status MatchStatus, err error) {

	hasher := sha256.New()
	if _, err = io.Copy(hasher, reader); err != nil {

		return Unknown, err
	}

	reqSha256Sum = strings.ToLower(reqSha256Sum)
	gotSha256Sum := fmt.Sprintf("%x", hasher.Sum(nil))
	if reqSha256Sum != gotSha256Sum {
		return NoMatch, nil
	}

	return Match, nil
}

// ValidateIntegrity validates the integrity of files on disk pertaining to each contract in the contract list.
// Integrity is validated by computing SHA256 sum of files on disk and comparing it with expected values in contract handlers.
func ValidateIntegrity(fs afero.Fs, contractList []Handler, contractsDir string) (corruptFiles, missingFiles []string, err error) {

	for _, contract := range contractList {

		fileTypeHashPairs := make(map[fileType]string)
		fileTypeHashPairs[golangFile] = contract.HashGoFile
		fileTypeHashPairs[solidityFile] = contract.HashSolFile
		fileTypeHashPairs[binRuntimeFile] = contract.HashBinRuntimeFile

		for fileType, fileHash := range fileTypeHashPairs {

			var matchResult MatchStatus
			var file afero.File
			var fileName string

			fileName, err = contract.getAbsFilepath(contractsDir, fileType)
			if err != nil {
				matchResult = Missing
				goto resultSwitch
			}

			file, err = fs.Open(fileName)
			if err != nil {
				if os.IsNotExist(err) {
					matchResult = Missing
					goto resultSwitch
				}
				return corruptFiles, missingFiles, err
			}

			matchResult, err = CheckSha256Sum(file, fileHash)

		resultSwitch:
			switch matchResult {
			case Match:
				continue
			case NoMatch:
				corruptFiles = append(corruptFiles, fileName)
			case Missing:
				missingFiles = append(missingFiles, fileName)
			}

		}
	}

	if len(missingFiles) != 0 || len(corruptFiles) != 0 {
		err = fmt.Errorf("corrupt/missing files found")
	}
	return corruptFiles, missingFiles, err
}
