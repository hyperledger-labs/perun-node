// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
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

// Package local implements an implementation of ID provider, where
// the peer IDs are stored in a YAML file stored locally on the disk.
//
// The complete list of peer IDs are loaded into an in-memory cache during
// initialization. The entries in the cache are indexed by both alias and
// off-chain address of the peer and can be using either of these as reference.
//
// Read, Write and Delete operations act only on the cache and do
// not affect the contents of the file.
//
// Latest state of cache can be updated to the file by explicitly calling
// UpdateStorage method. Normally this should be called before shutting down
// the node.
package local
