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

package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	gel "github.com/geldata/gel-go/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dsn string

var tests = []struct {
	description string
	directory   string
	args        []string
}{
	{
		description: "invoke edgeql-go without args",
		directory:   "testdata/no-args",
		args:        []string{},
	},
	{
		description: "invoke edgeql-go with -mixedcaps",
		directory:   "testdata/mixedcaps",
		args:        []string{"-mixedcaps"},
	},
	{
		description: "invoke edgeql-go with -pubfuncs",
		directory:   "testdata/pubfuncs",
		args:        []string{"-pubfuncs"},
	},
	{
		description: "invoke edgeql-go with -pubtypes",
		directory:   "testdata/pubtypes",
		args:        []string{"-pubtypes"},
	},
}

func TestMain(m *testing.M) {
	o := gel.TestClientOptions()
	pwd, ok := o.Password.Get()
	if !ok {
		log.Fatal("missing password")
	}
	dsn = fmt.Sprintf(
		"edgedb://%s:%s@%s:%d?tls_security=%s&tls_ca_file=%s",
		o.User,
		pwd,
		o.Host,
		o.Port,
		o.TLSOptions.SecurityMode,
		o.TLSOptions.CAFile,
	)
	os.Exit(m.Run())
}

func TestEdgeQLGo(t *testing.T) {
	for _, test := range tests {
		t.Run(test.description, runTest(test.directory, test.args))
	}
}

func runTest(dir string, args []string) func(*testing.T) {
	return func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "edgeql-go-*")
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tmpDir))
		}()

		t.Log("building edgeql-go")
		edgeqlGo := filepath.Join(tmpDir, "edgeql-go")
		run(t, ".", "go", "build", "-o", edgeqlGo)

		var wg sync.WaitGroup
		err = filepath.WalkDir(
			dir,
			func(src string, d fs.DirEntry, e error) error {
				require.NoError(t, e)
				if src == dir {
					return nil
				}

				dst := filepath.Join(tmpDir, strings.TrimPrefix(src, dir))
				if d.IsDir() {
					e = os.Mkdir(dst, os.ModePerm)
					require.NoError(t, e)
				} else {
					wg.Add(1)
					go func() {
						defer wg.Done()
						copyFile(t, dst, src)
					}()
				}
				return nil
			},
		)
		require.NoError(t, err)
		wg.Wait()

		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		for _, entry := range entries {
			if entry.Name() == "edgeql-go" {
				continue
			}

			t.Run(entry.Name(), func(t *testing.T) {
				projectDir := filepath.Join(tmpDir, entry.Name())
				run(t, projectDir, edgeqlGo, args...)
				run(t, projectDir, "go", "run", "./...")
				er := filepath.WalkDir(
					projectDir,
					func(f string, d fs.DirEntry, e error) error {
						require.NoError(t, e)
						if strings.HasSuffix(f, ".go.assert") {
							checkAssertFile(t, f)
						}
						if strings.HasSuffix(f, ".go") &&
							!strings.HasSuffix(f, "ignore.go") {
							checkGoFile(t, f)
						}
						return nil
					},
				)
				require.NoError(t, er)
			})
		}
	}
}

func checkAssertFile(t *testing.T, file string) {
	t.Helper()
	goFile := strings.TrimSuffix(file, ".assert")
	if assert.FileExistsf(t, goFile, "missing .go file for %s", file) {
		assertEqualFiles(t, file, goFile)
	}
}

func checkGoFile(t *testing.T, file string) {
	t.Helper()
	assertFile := file + ".assert"
	if assert.FileExistsf(t, assertFile,
		"missing .go.assert file for %s", file,
	) {
		assertEqualFiles(t, assertFile, file)
	}
}

func assertEqualFiles(t *testing.T, left, right string) {
	t.Helper()
	leftData, err := os.ReadFile(left)
	require.NoErrorf(t, err, "reading %s", left)

	rightData, err := os.ReadFile(right)
	require.NoErrorf(t, err, "reading %s", right)

	assert.Equal(t, string(leftData), string(rightData),
		"files are not equal: %s != %s", left, right,
	)
}

func copyFile(t *testing.T, to, from string) {
	toFd, err := os.Create(to)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, toFd.Close())
	}()

	fromFd, err := os.Open(from)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, fromFd.Close())
	}()

	_, err = io.Copy(toFd, fromFd)
	require.NoError(t, err)
}

func run(t *testing.T, dir, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), fmt.Sprintf("EDGEDB_DSN=%s", dsn))
	stdoutStderr, err := cmd.CombinedOutput()
	require.NoError(t, err, string(stdoutStderr))
}
