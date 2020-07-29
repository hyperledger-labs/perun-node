// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/direct-state-transfer/perun-node
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

// Package ethereum provides on-chain transaction backend and wallet backend
// for the ethereum blockchain platform. The actual implementation of the
// functionality is done in internal package. This implementation can be
// configured for both real and test uses and shared by this package and
// the the ethereumtest package.
//
// In addition to the intended functionality, this package is also structured
// to isolate all the imports from "go-ethereum" project and
// "go-perun/ethereum/backend" package in go-perun project, as the former is
// licensed under LGPL and the latter imports (and hence statically links)
// to code that is licensed under LGPL.
//
// In order to provide this isolation the exported methods in this packages
// use only those types defined in the root package of this project and in std lib.
// This restriction enables the other packages in perun-node to compile this package
// as plugin and load the symbols from it in runtime (using "plugin" library)
// without importing any package from "go-perun/backend/ethereum" or
// "go-ethereum" or this package.
// This enables the possibility to generate a perun-node binary that does not contain any
// components licensed under LGPL, yet use the functionality from such
// libraries by dynamically linking to them during runtime.
package ethereum
