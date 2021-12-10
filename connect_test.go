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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	ctx := context.Background()
	p, err := CreateClient(ctx, Options{
		Host:       opts.Host,
		Port:       opts.Port,
		User:       "user_with_password",
		Password:   NewOptionalStr("secret"),
		Database:   opts.Database,
		TLSOptions: opts.TLSOptions,
	})
	require.NoError(t, err)

	var result string
	err = p.QuerySingle(ctx, "SELECT 'It worked!';", &result)
	assert.NoError(t, err)
	assert.Equal(t, "It worked!", result)

	clientCopy := p.WithTxOptions(NewTxOptions())

	err = p.Close()
	assert.NoError(t, err)

	// A connection should not be closeable more than once.
	err = p.Close()
	msg := "edgedb.InterfaceError: client closed"
	assert.EqualError(t, err, msg)

	// Copied connections should not be closeable after another copy is closed.
	err = clientCopy.Close()
	assert.EqualError(t, err, msg)
}
