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

	"github.com/edgedb/edgedb-go/internal/cache"
)

// Conn is a single Conn to a server.
// Conn is not safe for concurrent use.
// Pool should be preferred over Conn for most use cases.
//
// Deprecated: use a Pool from Connect() or ConnectDSN()
type Conn struct {
	transactableConn
}

// ConnectOne establishes a connection to an EdgeDB server.
//
// Deprecated: use Connect() instead
func ConnectOne(ctx context.Context, opts Options) (*Conn, error) { // nolint:gocritic,lll
	return ConnectOneDSN(ctx, "", opts)
}

// ConnectOneDSN establishes a connection to an EdgeDB server.
//
// dsn is either an instance name
// https://www.edgedb.com/docs/clients/00_python/instances/#edgedb-instances
// or it specifies a single string in the following format:
//
//     edgedb://user:password@host:port/database?option=value.
//
// The following options are recognized: host, port, user, database, password.
//
// Deprecated: use ConnectDSN() instead
func ConnectOneDSN(
	ctx context.Context,
	dsn string,
	opts Options, // nolint:gocritic
) (*Conn, error) {
	config, err := parseConnectDSNAndArgs(dsn, &opts)
	if err != nil {
		return nil, err
	}

	conn := &Conn{transactableConn{
		txOpts: NewTxOptions(),
		reconnectingConn: &reconnectingConn{
			cfg: config,
			cacheCollection: cacheCollection{
				serverSettings:    config.serverSettings,
				typeIDCache:       cache.New(1_000),
				inCodecCache:      cache.New(1_000),
				outCodecCache:     cache.New(1_000),
				capabilitiesCache: cache.New(1_000),
			},
		},
	}}

	if err := conn.reconnect(ctx, false); err != nil {
		return nil, err
	}

	return conn, nil
}
