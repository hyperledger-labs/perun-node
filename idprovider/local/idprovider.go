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

package local

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/hyperledger-labs/perun-node"
)

// Provider represents a contacts provider that provides access to contacts stored in a yaml file.
//
// It generates a cache of all contacts in the yaml file during initialization. Read, Write and Delete
// operations act only on the cached list of contacts and do not update the yaml file.
// The changes in cache can be updated to the yaml file by explicitly calling UpdateStorage method.
//
// It also stores an instance of wallet backend that will be used or decoding address strings.
type Provider struct {
	*contactsCache

	contactsFilePath string
}

// New returns an instance of contacts provider to access the contacts in the given yaml file.
//
// All the contacts are cached in memory during initialization and Read, Write, Delete operations
// affect only the cache. The changes are updated to the yaml file only when UpdateStorage method
// is explicitly called. There is no mechanism to reload the cache if the yaml file is updated.
//
// Backend is used for decoding the address strings during initialization.
func New(filePath string, backend perun.WalletBackend) (*Provider, error) {
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint: errcheck, gosec  // safe to defer f.Close() for files opened in read mode.

	cache := make(map[string]perun.Peer)
	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(&cache); err != nil && err != io.EOF {
		return nil, err
	}

	contactsCache, err := newContactsCache(cache, backend)
	if err != nil {
		return nil, err
	}
	return &Provider{
		contactsCache:    contactsCache,
		contactsFilePath: filePath,
	}, nil
}

// UpdateStorage writes the latest state of contacts cache to the yaml file.
func (c *Provider) UpdateStorage() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	f, err := os.Create(c.contactsFilePath)
	if err != nil {
		return errors.Wrap(err, "opening contacts file for writing")
	}
	defer func() {
		if fCloseErr := f.Close(); fCloseErr != nil {
			err = fmt.Errorf("%w; and error closing file - %s", err, fCloseErr.Error())
		}
	}()

	encoder := yaml.NewEncoder(f)
	if err = encoder.Encode(c.peersByAlias); err != nil {
		return errors.Wrap(err, "encoding data as yaml")
	}
	err = errors.Wrap(encoder.Close(), "closing encoder")
	// receive the error in "err" before returning to ensure file close error is captured.
	return err
}
