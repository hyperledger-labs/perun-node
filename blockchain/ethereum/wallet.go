// Copyright (c) 2020 - for information on the respective copyright owner
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

package ethereum

import (
	"github.com/direct-state-transfer/perun-node"
	"github.com/direct-state-transfer/perun-node/blockchain/ethereum/internal"
)

// NewWalletBackend initializes an ethereum specific wallet backend.
//
// The function signature uses only types defined in the root package of this project and types from std lib.
// This enables the function to be loaded as symbol without importing this package when it is compiled as plugin.
func NewWalletBackend() perun.WalletBackend {
	return &internal.WalletBackend{EncParams: internal.ScryptParams{
		N: internal.StandardScryptN,
		P: internal.StandardScryptP,
	}}
}
