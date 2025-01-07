// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
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

//go:build windows
// +build windows

package gel

import (
	"crypto/x509"
	"path/filepath"

	"github.com/certifi/gocertifi"
	"golang.org/x/sys/windows"
)

func getSystemCertPool() (*x509.CertPool, error) {
	// x509.SystemCertPool() doesn't work on Windows.
	// https://github.com/golang/go/issues/16736
	return gocertifi.CACerts()
}

func configDirOSSpecific() (string, error) {
	dir, err := windows.KnownFolderPath(
		windows.FOLDERID_LocalAppData, windows.KF_FLAG_DEFAULT)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "EdgeDB", "config"), nil
}

func device(dir string) (int, error) {
	return 0, nil
}
