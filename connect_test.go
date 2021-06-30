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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	ctx := context.Background()
	conn, err := ConnectOne(ctx, Options{
		Hosts:             opts.Hosts,
		Ports:             opts.Ports,
		User:              "user_with_password",
		Password:          "secret",
		Database:          opts.Database,
		TLSCAFile:         opts.TLSCAFile,
		TLSVerifyHostname: opts.TLSVerifyHostname,
	})
	require.Nil(t, err, "unexpected error: %v", err)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	var result string
	err = conn.QueryOne(ctx, "SELECT 'It worked!';", &result)
	cancel()

	require.Nil(t, err, "unexpected error: %v", err)
	assert.Equal(t, "It worked!", result)
}
