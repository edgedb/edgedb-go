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

package edgedb

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

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
					addrs:              []*dialArgs{{"tcp", "localhost:5656"}},
					user:               "user",
					database:           "edgedb",
					serverSettings:     map[string]string{},
					waitUntilAvailable: 30 * time.Second,
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
					addrs:              []*dialArgs{{"tcp", "host:123"}},
					user:               "user",
					password:           "passw",
					database:           "testdb",
					serverSettings:     map[string]string{},
					waitUntilAvailable: 30 * time.Second,
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
				Password: NewOptionalStr("passw2"),
				Database: "db2",
			},
			expected: Result{
				cfg: connConfig{
					addrs:              []*dialArgs{{"tcp", "host2:456"}},
					user:               "user2",
					password:           "passw2",
					database:           "db2",
					serverSettings:     map[string]string{},
					waitUntilAvailable: 30 * time.Second,
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
				// Host:           "host2",
				// Port:           456,
				User:           "user2",
				Password:       NewOptionalStr("passw2"),
				Database:       "db2",
				ServerSettings: map[string]string{"ssl": "False"},
			},
			expected: Result{
				cfg: connConfig{
					addrs:              []*dialArgs{{"tcp", "localhost:5656"}},
					user:               "user2",
					password:           "passw2",
					database:           "db2",
					serverSettings:     map[string]string{"ssl": "False"},
					waitUntilAvailable: 30 * time.Second,
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
					addrs:              []*dialArgs{{"tcp", "localhost:5555"}},
					user:               "user3",
					password:           "123123",
					database:           "abcdef",
					serverSettings:     map[string]string{},
					waitUntilAvailable: 30 * time.Second,
				},
			},
		},
		{
			name: "DSN only",
			dsn:  "edgedb://user3:123123@localhost:5555/abcdef",
			expected: Result{
				cfg: connConfig{
					addrs:              []*dialArgs{{"tcp", "localhost:5555"}},
					user:               "user3",
					password:           "123123",
					database:           "abcdef",
					serverSettings:     map[string]string{},
					waitUntilAvailable: 30 * time.Second,
				},
			},
		},
		{
			name: "DSN with multiple hosts",
			dsn:  "edgedb://user@host1,host2/db",
			expected: Result{
				err: &configurationError{},
				errMessage: `edgedb.ConfigurationError: ` +
					`invalid host: "host1,host2"`,
			},
		},
		{
			name: "DSN with multiple hosts and ports",
			dsn:  "edgedb://user@host1:1111,host2:2222/db",
			expected: Result{
				err: &configurationError{},
				errMessage: `edgedb.ConfigurationError: ` +
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
				errMessage: `edgedb.ConfigurationError: ` +
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
				errMessage: `edgedb.ConfigurationError: ` +
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
				errMessage: `edgedb.ConfigurationError: ` +
					`Cannot have more than one of the following ` +
					`connection options: dsn, CredentialsFile, or Host/Port`,
			},
		},
		{
			name: "DSN with server settings",
			dsn: "edgedb://?param=123&host=testhost&user=testuser" +
				"&port=2222&database=testdb",
			opts: Options{
				// Host:     "127.0.0.1",
				// Port:     888,
				User:     "me",
				Password: NewOptionalStr("ask"),
				Database: "db",
			},
			expected: Result{
				cfg: connConfig{
					addrs: []*dialArgs{
						{"tcp", "testhost:2222"},
					},
					serverSettings:     map[string]string{"param": "123"},
					user:               "me",
					password:           "ask",
					database:           "db",
					waitUntilAvailable: 30 * time.Second,
				},
			},
		},
		{
			name: "DSN and options server settings are merged",
			dsn: "edgedb://?param=123&host=testhost&user=testuser" +
				"&port=2222&database=testdb",
			opts: Options{
				// Host:           "127.0.0.1",
				// Port:           888,
				User:           "me",
				Password:       NewOptionalStr("ask"),
				Database:       "db",
				ServerSettings: map[string]string{"aa": "bb"},
			},
			expected: Result{
				cfg: connConfig{
					addrs: []*dialArgs{
						{"tcp", "testhost:2222"},
					},
					serverSettings: map[string]string{
						"aa":    "bb",
						"param": "123",
					},
					user:               "me",
					password:           "ask",
					database:           "db",
					waitUntilAvailable: 30 * time.Second,
				},
			},
		},
		{
			name: "DSN with unix socket",
			dsn:  "edgedb:///dbname?host=/unix_sock/test&user=spam",
			expected: Result{
				err: &configurationError{},
				errMessage: "edgedb.ConfigurationError: " +
					`unix socket paths not supported`,
			},
		},
		{
			name: "DSN requires edgedb scheme",
			dsn:  "pq:///dbname?host=/unix_sock/test&user=spam",
			expected: Result{
				err: &configurationError{},
				errMessage: "edgedb.ConfigurationError: " +
					`invalid DSN: scheme is expected to be "edgedb", got "pq"`,
			},
		},
		{
			name: "DSN query parameter with unix socket",
			dsn:  "edgedb://user@?port=56226&host=%2Ftmp",
			expected: Result{
				err: &configurationError{},
				errMessage: "edgedb.ConfigurationError: " +
					`unix socket paths not supported`,
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			cleanup := setenvmap(c.env)
			defer cleanup()

			config, err := parseConnectDSNAndArgs(c.dsn, &c.opts)

			if c.expected.err != nil {
				require.EqualError(t, err, c.expected.errMessage)
				require.True(t, errors.As(err, &c.expected.err))
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				// tlsConfigs cannot be compared reliably
				config.tlsConfig = nil
				assert.Equal(t, c.expected.cfg, *config)
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
	assert.EqualError(
		t,
		err,
		"edgedb.ClientConnectionFailedTemporarilyError: "+
			"dial tcp: lookup invalid.example.org: no such host",
	)

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
