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

//go:build darwin
// +build darwin

package edgedb

import (
	"crypto/x509"
	"os"
	"path"
	"syscall"
)

func getSystemCertPool() (*x509.CertPool, error) {
	return x509.SystemCertPool()
}

func configDirOSSpecific() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(dir, "Library", "Application Support", "edgedb"), nil
}

func device(dir string) (int, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return 0, err
	}

	return int(stat.Sys().(*syscall.Stat_t).Dev), nil
}