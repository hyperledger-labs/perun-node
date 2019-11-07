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

//go:generate echo -e "\\e[01;31mensure if solc Version is 0.4.24 before running go generate\\e[0m"
//go:generate solc --bin-runtime --optimize LibSignatures.sol --overwrite -o ./
//go:generate solc --bin-runtime --optimize VPC.sol --overwrite -o ./
//go:generate solc --bin-runtime --optimize MSContract.sol --overwrite -o ./
//go:generate abigen --pkg contract --sol MSContract.sol --out ../MSContract.go
//go:generate abigen --pkg contract --sol LibSignatures.sol --out ../LibSignatures.go
//go:generate echo "package contract" > ../VPC.go

//VPC is imported in MSContract.
//Hence, go bindings for the same will be available in MSContract.go
//Empty files created to pass hash validation

// *** IMPORTANT ***
// Delete the part corresponding to ILibSignatures in LibSignatures.go
