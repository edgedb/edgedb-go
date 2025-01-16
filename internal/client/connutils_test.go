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

package gel

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/geldata/gel-go/internal/geltypes"
	"github.com/geldata/gel-go/internal/snc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setenvmap(m map[string]string) func() {
	funcs := make([]func(), 0, len(m))
	for key, val := range m {
		funcs = append(funcs, setenv(key, val))
	}

	return func() {
		for _, fn := range funcs {
			fn()
		}
	}
}

func setenv(key, val string) func() {
	old, ok := os.LookupEnv(key)

	err := os.Setenv(key, val)
	if err != nil {
		panic(err)
	}

	if ok {
		return func() {
			err = os.Setenv(key, old)
			if err != nil {
				panic(err)
			}
		}
	}

	return func() {
		err = os.Unsetenv(key)
		if err != nil {
			panic(err)
		}
	}
}

func newServerSettingValues(settings map[string][]byte) *snc.ServerSettings {
	s := snc.NewServerSettings()
	for k, v := range settings {
		s.Set(k, v)
	}
	return s
}

func TestConUtils(t *testing.T) {
	type Result struct {
		cfg        connConfig
		err        error
		errMessage string
	}

	tests := []struct {
		name     string
		env      map[string]string
		dsn      string
		opts     Options
		expected Result
	}{
		{
			name: "host and user options",
			opts: Options{
				User: "user",
				Host: "localhost",
			},
			expected: Result{
				cfg: connConfig{
					addr:               dialArgs{"tcp", "localhost:5656"},
					user:               "user",
					database:           "edgedb",
					branch:             "__default__",
					serverSettings:     snc.NewServerSettings(),
					waitUntilAvailable: 30 * time.Second,
					tlsSecurity:        "strict",
				},
			},
		},
		{
			name: "all environment variables",
			env: map[string]string{
				"EDGEDB_USER":     "user",
				"EDGEDB_DATABASE": "testdb",
				"EDGEDB_PASSWORD": "passw",
				"EDGEDB_HOST":     "host",
				"EDGEDB_PORT":     "123",
			},
			expected: Result{
				cfg: connConfig{
					addr:               dialArgs{"tcp", "host:123"},
					user:               "user",
					password:           "passw",
					database:           "testdb",
					branch:             "testdb",
					serverSettings:     snc.NewServerSettings(),
					waitUntilAvailable: 30 * time.Second,
					tlsSecurity:        "strict",
				},
			},
		},
		{
			name: "options are used before environment variables",
			env: map[string]string{
				"EDGEDB_USER":     "user",
				"EDGEDB_DATABASE": "testdb",
				"EDGEDB_PASSWORD": "passw",
				"EDGEDB_HOST":     "host",
				"EDGEDB_PORT":     "123",
			},
			opts: Options{
				Host:     "host2",
				Port:     456,
				User:     "user2",
				Password: geltypes.NewOptionalStr("passw2"),
				Database: "db2",
			},
			expected: Result{
				cfg: connConfig{
					addr:               dialArgs{"tcp", "host2:456"},
					user:               "user2",
					password:           "passw2",
					database:           "db2",
					branch:             "db2",
					serverSettings:     snc.NewServerSettings(),
					waitUntilAvailable: 30 * time.Second,
					tlsSecurity:        "strict",
				},
			},
		},
		{
			name: "options are used before DSN string",
			env: map[string]string{
				"EDGEDB_USER":     "user",
				"EDGEDB_DATABASE": "testdb",
				"EDGEDB_PASSWORD": "passw",
				"EDGEDB_HOST":     "host",
				"EDGEDB_PORT":     "123",
				"PGSSLMODE":       "prefer",
			},
			dsn: "edgedb://user3:123123@localhost/abcdef",
			opts: Options{
				User:           "user2",
				Password:       geltypes.NewOptionalStr("passw2"),
				Database:       "db2",
				ServerSettings: map[string][]byte{"ssl": []byte("False")},
			},
			expected: Result{
				cfg: connConfig{
					addr:     dialArgs{"tcp", "localhost:5656"},
					user:     "user2",
					password: "passw2",
					database: "db2",
					branch:   "db2",
					serverSettings: newServerSettingValues(map[string][]byte{
						"ssl": []byte("False"),
					}),
					waitUntilAvailable: 30 * time.Second,
					tlsSecurity:        "strict",
				},
			},
		},
		{
			name: "DSN is used before environment variables",
			env: map[string]string{
				"EDGEDB_USER":     "user",
				"EDGEDB_DATABASE": "testdb",
				"EDGEDB_PASSWORD": "passw",
				"EDGEDB_HOST":     "host",
				"EDGEDB_PORT":     "123",
			},
			dsn: "edgedb://user3:123123@localhost:5555/abcdef",
			expected: Result{
				cfg: connConfig{
					addr:               dialArgs{"tcp", "localhost:5555"},
					user:               "user3",
					password:           "123123",
					database:           "abcdef",
					branch:             "abcdef",
					serverSettings:     snc.NewServerSettings(),
					waitUntilAvailable: 30 * time.Second,
					tlsSecurity:        "strict",
				},
			},
		},
		{
			name: "DSN only",
			dsn:  "edgedb://user3:123123@localhost:5555/abcdef",
			expected: Result{
				cfg: connConfig{
					addr:               dialArgs{"tcp", "localhost:5555"},
					user:               "user3",
					password:           "123123",
					database:           "abcdef",
					branch:             "abcdef",
					serverSettings:     snc.NewServerSettings(),
					waitUntilAvailable: 30 * time.Second,
					tlsSecurity:        "strict",
				},
			},
		},
		{
			name: "DSN with multiple hosts",
			dsn:  "edgedb://user@host1,host2/db",
			expected: Result{
				err: &configurationError{},
				errMessage: `gel.ConfigurationError: invalid DSN: ` +
					`invalid host: "host1,host2"`,
			},
		},
		{
			name: "DSN with multiple hosts and ports",
			dsn:  "edgedb://user@host1:1111,host2:2222/db",
			expected: Result{
				err: &configurationError{},
				errMessage: `gel.ConfigurationError: invalid DSN: ` +
					`invalid host: "host1:1111,host2"`,
			},
		},
		{
			name: "environment variables with multiple hosts and ports",
			env: map[string]string{
				"EDGEDB_HOST": "host1:1111,host2:2222",
				"EDGEDB_USER": "foo",
			},
			dsn: "",
			expected: Result{
				err: &configurationError{},
				errMessage: `gel.ConfigurationError: ` +
					`invalid host: "host1:1111,host2:2222"`,
			},
		},
		{
			name: "query parameters with multiple hosts and ports",
			env: map[string]string{
				"EDGEDB_USER": "foo",
			},
			dsn: "edgedb:///db?host=host1:1111,host2:2222",
			expected: Result{
				err: &configurationError{},
				errMessage: `gel.ConfigurationError: invalid DSN: ` +
					`invalid host: "host1:1111,host2:2222"`,
			},
		},
		{
			name: "multiple compound options",
			env: map[string]string{
				"EDGEDB_USER": "foo",
			},
			dsn: "edgedb:///db",
			opts: Options{
				Host: "host1,host2",
			},
			expected: Result{
				err: &configurationError{},
				errMessage: `gel.ConfigurationError: ` +
					`mutually exclusive connection options specified: ` +
					`dsn and gel.Options.Host`,
			},
		},
		{
			name: "DSN with server settings",
			dsn: "edgedb://?param=123&host=testhost&user=testuser" +
				"&port=2222&database=testdb",
			opts: Options{
				User:     "me",
				Password: geltypes.NewOptionalStr("ask"),
				Database: "db",
			},
			expected: Result{
				cfg: connConfig{
					addr: dialArgs{"tcp", "testhost:2222"},
					serverSettings: newServerSettingValues(map[string][]byte{
						"param": []byte("123"),
					}),
					user:               "me",
					password:           "ask",
					database:           "db",
					branch:             "db",
					waitUntilAvailable: 30 * time.Second,
					tlsSecurity:        "strict",
				},
			},
		},
		{
			name: "DSN and options server settings are merged",
			dsn: "edgedb://?param=123&host=testhost&user=testuser" +
				"&port=2222&database=testdb",
			opts: Options{
				User:           "me",
				Password:       geltypes.NewOptionalStr("ask"),
				Database:       "db",
				ServerSettings: map[string][]byte{"aa": []byte("bb")},
			},
			expected: Result{
				cfg: connConfig{
					addr: dialArgs{"tcp", "testhost:2222"},
					serverSettings: newServerSettingValues(map[string][]byte{
						"aa":    []byte("bb"),
						"param": []byte("123"),
					}),
					user:               "me",
					password:           "ask",
					database:           "db",
					branch:             "db",
					waitUntilAvailable: 30 * time.Second,
					tlsSecurity:        "strict",
				},
			},
		},
		{
			name: "DSN with unix socket",
			dsn:  "edgedb:///dbname?host=/unix_sock/test&user=spam",
			expected: Result{
				err: &configurationError{},
				errMessage: `gel.ConfigurationError: invalid DSN: ` +
					`invalid host: unix socket paths not supported, ` +
					`got "/unix_sock/test"`,
			},
		},
		{
			name: "DSN requires edgedb scheme",
			dsn:  "pq:///dbname?host=/unix_sock/test&user=spam",
			expected: Result{
				err: &configurationError{},
				errMessage: "gel.ConfigurationError: " +
					`invalid DSN: scheme is expected to be "gel", got "pq"`,
			},
		},
		{
			name: "DSN query parameter with unix socket",
			dsn:  "edgedb://user@?port=56226&host=%2Ftmp",
			expected: Result{
				err: &configurationError{},
				errMessage: `gel.ConfigurationError: invalid DSN: ` +
					`invalid host: unix socket paths not supported, ` +
					`got "/tmp"`,
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			cleanup := setenvmap(c.env)
			defer cleanup()

			config, err := parseConnectDSNAndArgs(
				c.dsn, &c.opts, newCfgPaths())

			if c.expected.err != nil {
				require.EqualError(t, err, c.expected.errMessage)
				require.True(t, errors.As(err, interface{}(&c.expected.err)))
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				assert.Equal(t, c.expected.cfg, *config)
			}
		})
	}
}

var testcaseErrorMapping = map[string]string{
	"credentials_file_not_found": "cannot read credentials",
	"project_not_initialised":    "project is not initialized",
	"no_options_or_toml": "no `gel.toml` found and no connection options " +
		"specified either",
	"invalid_credentials_file":     "cannot parse credentials",
	"invalid_dsn_or_instance_name": "invalid DSN|invalid instance name",
	"invalid_instance_name":        "invalid instance name",
	"invalid_dsn":                  "invalid DSN",
	"unix_socket_unsupported":      "unix socket paths not supported",
	"invalid_port":                 "invalid port",
	"invalid_host":                 "invalid host",
	"invalid_user":                 "invalid user",
	"invalid_database":             "invalid database",
	"exclusive_options":            "mutually exclusive options",
	"multiple_compound_opts":       "mutually exclusive connection options",
	"multiple_compound_env":        "mutually exclusive environment variables",
	"env_not_found":                "environment variable .* is not set",
	"file_not_found": "no such file or directory|" +
		"cannot find the (?:file|path) specified",
	"invalid_tls_security": "invalid TLSSecurity value|tls_verify_hostname" +
		"=.* and tls_security=.* are incompatible" +
		"|tls_security must be set to strict",
	"secret_key_not_found": "Cannot connect to cloud instances " +
		"without secret key",
	"invalid_secret_key": "Invalid secret key",
}

var testcaseWarningMapping = map[string]*regexp.Regexp{
	"docker_tcp_port": regexp.MustCompile(
		`ignoring (EDGEDB|GEL)_PORT in 'tcp:\/\/host:port' format`),
	"gel_and_edgedb": regexp.MustCompile(
		`Both GEL_\w+ and EDGEDB_\w+ are set. EDGEDB_\w+ will be ignored.`),
}

func getStr(t *testing.T, lookup map[string]interface{}, key string) string {
	if lookup[key] == nil {
		return ""
	}

	str, ok := lookup[key].(string)
	if !ok {
		t.Skipf("%v is not a string", key)
	} else if str == "" {
		t.Skipf("%v is an empty string", key)
	}

	return str
}

func getBytes(t *testing.T, lookup map[string]interface{}, key string) []byte {
	str := getStr(t, lookup, key)
	if str == "" {
		return nil
	}

	return []byte(str)
}

func getDuration(
	t *testing.T,
	lookup map[string]interface{},
	key string,
) time.Duration {
	val, ok := lookup[key]
	require.True(t, ok, "%q is missing", key)

	str, ok := val.(string)
	require.True(t, ok, "%q should be a string", key)

	dur, err := geltypes.ParseDuration(str)
	require.NoError(t, err, "could not parse %q duration", key)

	return time.Duration(1_000 * dur)
}

func configureFileSystem(
	tmpDir string,
	cfg map[string]interface{},
	paths *cfgPaths,
) error {
	var err error
	paths.testDir = tmpDir
	paths.cfgDir = filepath.Join(tmpDir, "home", "edgedb", ".config", "edgedb")

	if cwd, ok := cfg["cwd"]; ok {
		paths.cwdErr = nil
		paths.cwd = filepath.Join(tmpDir, cwd.(string))
		err = os.MkdirAll(paths.cwd, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if home, ok := cfg["homedir"]; ok {
		paths.cfgDirErr = nil
		paths.cfgDir = filepath.Join(
			tmpDir, home.(string), ".config", "edgedb")
		err = os.MkdirAll(paths.cfgDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if files, ok := cfg["files"]; ok {
		for file, data := range files.(map[string]interface{}) {
			switch x := data.(type) {
			case string:
				err = createFile(tmpDir, file, x)
			case map[string]interface{}:
				err = createProjectDir(tmpDir, file, x)
			default:
				err = fmt.Errorf("unexpected data type %T", data)
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createProjectDir(
	tmpDir, dir string,
	contents map[string]interface{},
) error {
	path := filepath.Join(tmpDir, contents["project-path"].(string))
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" && !strings.HasPrefix(path, `\\`) {
		path = `\\?\` + path
	}

	hash := fmt.Sprintf("%x", sha1.Sum([]byte(path)))
	dir = strings.Replace(dir, "${HASH}", hash, 1)

	for name, content := range contents {
		err = createFile(tmpDir, filepath.Join(dir, name), content.(string))
		if err != nil {
			return err
		}
	}

	return nil
}

func createFile(tmpDir, file, data string) error {
	file = filepath.Join(tmpDir, file)
	err := os.MkdirAll(filepath.Dir(file), os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(file, []byte(data), 0644)
}

func TestConnectionParameterResolution(t *testing.T) {
	data, err := os.ReadFile(
		"../../shared-client-testcases/connection_testcases.json",
	)
	require.NoError(t, err, "Failed to read 'connection_testcases.json'\n"+
		"Is the 'shared-client-testcases' submodule initialised? "+
		"Try running 'git submodule update --init'.")

	var testcases []map[string]interface{}
	err = json.Unmarshal(data, &testcases)
	require.NoError(t, err)

	for i, testcase := range testcases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if _, ok := testcase["platform"]; ok {
				t.Skip("platform specific tests not supported")
			}
			tmpDir, err := os.MkdirTemp(os.TempDir(), "gel-go-tests")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir) // nolint:errcheck
			paths := newCfgPaths()
			if fs, ok := testcase["fs"]; ok {
				err = configureFileSystem(
					tmpDir, fs.(map[string]interface{}), paths)
				require.NoError(t, err)
			}
			env := make(map[string]string)
			if testcase["env"] != nil {
				testcaseEnv := testcase["env"].(map[string]interface{})
				for k, v := range testcaseEnv {
					if strings.HasSuffix(k, "_FILE") {
						v = filepath.Join(tmpDir, v.(string))
					}
					env[k] = v.(string)
				}
			}
			if len(env) > 0 {
				cleanup := setenvmap(env)
				defer cleanup()
			}

			var dsn string
			var options Options

			if opts, ok := testcase["opts"].(map[string]interface{}); ok {
				if inst, ok := opts["instance"]; ok {
					opts["dsn"] = inst
				}

				dsn = getStr(t, opts, "dsn")
				dsn = strings.ReplaceAll(dsn, "_file=/", "_file="+tmpDir+"/")
				file := getStr(t, opts, "credentialsFile")
				if file != "" {
					options.CredentialsFile = filepath.Join(tmpDir, file)
				}
				options.Credentials = getBytes(t, opts, "credentials")
				options.Host = getStr(t, opts, "host")
				if opts["port"] != nil {
					options.Port, _ = opts["port"].(int)
					if options.Port == 0 {
						t.Skip("unusable port value")
					}
				}
				options.Database = getStr(t, opts, "database")
				options.Branch = getStr(t, opts, "branch")
				options.User = getStr(t, opts, "user")
				if opts["password"] != nil {
					options.Password.Set(opts["password"].(string))
				}
				file = getStr(t, opts, "tlsCAFile")
				if file != "" {
					options.TLSOptions.CAFile = filepath.Join(tmpDir, file)
				}
				options.TLSOptions.CA = getBytes(t, opts, "tlsCA")
				options.TLSOptions.SecurityMode = TLSSecurityMode(
					getStr(t, opts, "tlsSecurity"))
				options.TLSOptions.ServerName = getStr(
					t, opts, "tlsServerName")
				if opts["serverSettings"] != nil {
					ss := opts["serverSettings"].(map[string]interface{})
					options.ServerSettings = make(map[string][]byte, len(ss))
					for k, v := range ss {
						options.ServerSettings[k] = []byte(v.(string))
					}
				}
				if opts["waitUntilAvailable"] != nil {
					options.WaitUntilAvailable = getDuration(
						t,
						opts,
						"waitUntilAvailable",
					)
				}

				options.SecretKey = getStr(t, opts, "secretKey")
			}

			expectedResult := connConfig{
				serverSettings:     snc.NewServerSettings(),
				waitUntilAvailable: 30 * time.Second,
			}

			if testcase["result"] != nil {
				res := testcase["result"].(map[string]interface{})
				addr := res["address"].([]interface{})

				expectedResult.addr = dialArgs{
					network: "tcp",
					address: fmt.Sprintf("%v:%v", addr[0], addr[1]),
				}
				expectedResult.database = res["database"].(string)
				expectedResult.branch = res["branch"].(string)
				expectedResult.user = res["user"].(string)
				if res["password"] != nil {
					expectedResult.password = res["password"].(string)
				}

				expectedResult.tlsSecurity = res["tlsSecurity"].(string)
				if data := res["tlsCAData"]; data != nil {
					expectedResult.tlsCAData = []byte(data.(string))
				}

				if key := res["secretKey"]; key != nil {
					expectedResult.secretKey = key.(string)
				}

				if key := res["tlsServerName"]; key != nil {
					expectedResult.tlsServerName = key.(string)
				}

				ss := res["serverSettings"].(map[string]interface{})
				for k, v := range ss {
					expectedResult.serverSettings.Set(k, []byte(v.(string)))
				}

				expectedResult.waitUntilAvailable = getDuration(
					t,
					res,
					"waitUntilAvailable",
				)
			}

			var testlogs bytes.Buffer
			log.SetOutput(&testlogs)
			defer log.SetOutput(os.Stderr)
			config, err := parseConnectDSNAndArgs(dsn, &options, paths)

			if testcase["warnings"] != nil {
				for _, warning := range testcase["warnings"].([]any) {
					regex, ok := testcaseWarningMapping[warning.(string)]
					if !ok {
						assert.Truef(
							t,
							false,
							"unexpected warning found "+
								"in shared-client-testcases: %v",
							warning,
						)
					} else {
						assert.Truef(
							t,
							regex.Match(testlogs.Bytes()),
							"no match for regex %q found in %q",
							regex.String(),
							testlogs.String(),
						)
					}
				}
			}

			if testcase["error"] != nil {
				errType := &configurationError{}
				require.IsType(t, errType, err)
				e := testcase["error"].(map[string]interface{})
				id := e["type"].(string)
				expected, ok := testcaseErrorMapping[id]
				require.Truef(t, ok, "unknown error type: %q", id)
				require.Regexp(t, expected, err.Error())
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				assert.Equal(t, expectedResult, *config)
			}
		})
	}
}

func TestConnectTimeout(t *testing.T) {
	ctx := context.Background()
	p, err := CreateClient(ctx, Options{
		Host:               opts.Host,
		Port:               opts.Port,
		User:               opts.User,
		Password:           opts.Password,
		Database:           opts.Database,
		ConnectTimeout:     2 * time.Nanosecond,
		WaitUntilAvailable: 1 * time.Nanosecond,
	})

	if p != nil {
		err = p.EnsureConnected(ctx)
		_ = p.Close()
	}

	require.NotNil(t, err, "connection didn't timeout")

	var edbErr Error

	require.True(t, errors.As(err, &edbErr), "wrong error: %v", err)
	assert.True(
		t,
		edbErr.Category(ClientConnectionTimeoutError),
		"wrong error: %v",
		err,
	)
}

func TestConnectRefused(t *testing.T) {
	ctx := context.Background()
	p, err := CreateClient(ctx, Options{
		Host:               "localhost",
		Port:               23456,
		WaitUntilAvailable: 1 * time.Nanosecond,
	})

	if p != nil {
		err = p.EnsureConnected(ctx)
		_ = p.Close()
	}

	require.NotNil(t, err, "connection wasn't refused")

	msg := "wrong error: " + err.Error()
	var edbErr Error
	require.True(t, errors.As(err, &edbErr), msg)
	assert.True(
		t,
		edbErr.Category(ClientConnectionFailedError),
		msg,
	)
}

func TestConnectInvalidName(t *testing.T) {
	ctx := context.Background()
	p, err := CreateClient(ctx, Options{
		Host:               "invalid.example.org",
		Port:               23456,
		WaitUntilAvailable: 1 * time.Nanosecond,
	})

	if p != nil {
		err = p.EnsureConnected(ctx)
		_ = p.Close()
	}

	require.NotNil(t, err, "name was resolved")

	var edbErr Error
	require.True(t, errors.As(err, &edbErr), "wrong error: %v", err)
	assert.True(
		t,
		edbErr.Category(ClientConnectionFailedTemporarilyError),
		"wrong error: %v",
		err,
	)
	// Match lookup error agnostic to OS. Examples:
	// dial tcp: lookup invalid.example.org: no such host
	// dial tcp: lookup invalid.example.org on 127.0.0.1:53: no such host
	assert.Contains(t, err.Error(),
		"gel.ClientConnectionFailedTemporarilyError: "+
			"dial tcp: lookup invalid.example.org")
	assert.Contains(t, err.Error(), "no such host")

	var errNotFound *net.DNSError
	assert.True(t, errors.As(err, &errNotFound))
}

func TestConnectRefusedUnixSocket(t *testing.T) {
	ctx := context.Background()
	p, err := CreateClient(ctx, Options{
		Host:               "/tmp/non-existent",
		WaitUntilAvailable: 1 * time.Nanosecond,
	})

	if p != nil {
		err = p.EnsureConnected(ctx)
		_ = p.Close()
	}

	require.NotNil(t, err, "connection wasn't refused")

	var edbErr Error
	require.True(t, errors.As(err, &edbErr), "wrong error: %v", err)
	assert.True(
		t,
		edbErr.Category(ConfigurationError),
		"wrong error: %v",
		err,
	)
}
