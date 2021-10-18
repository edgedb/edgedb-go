// This source file is part of the EdgeDB open source project.
//
// Copyright 2020-present EdgeDB Inc. and the EdgeDB authors.
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

//go:build !windows && !darwin
// +build !windows,!darwin

package edgedb

import (
	"crypto/x509"
	"os"
	"path"
)

func getSystemCertPool() (*x509.CertPool, error) {
	return x509.SystemCertPool()
}

func configDir() (string, error) {
	dir, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		dir = "."
	}

	if !path.IsAbs(dir) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		dir = path.Join(homeDir, ".config")
	}

	return path.Join(dir, "edgedb"), nil
}
