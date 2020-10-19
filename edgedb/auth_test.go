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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	var host string
	if server.admin {
		host = "localhost"
	} else {
		host = server.Host
	}

	conn, err := Connect(&Options{
		Host:     host,
		Port:     server.Port,
		User:     "user_with_password",
		Password: "secret",
		Database: server.Database,
	})
	assert.Nil(t, err)

	result, err := conn.QueryOneJSON("SELECT 'It worked!';")
	assert.Nil(t, err)
	assert.Equal(t, `"It worked!"`, string(result))
}
