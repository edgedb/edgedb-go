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

func TestParseHost(t *testing.T) {
	opts := DSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, "localhost", opts.Host)
}

func TestParsePort(t *testing.T) {
	opts := DSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, 5656, opts.Port)
}

func TestParseUser(t *testing.T) {
	opts := DSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, "me", opts.User)
}

func TestParseDatabase(t *testing.T) {
	opts := DSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, "somedb", opts.Database)
}

func TestParsePassword(t *testing.T) {
	opts := DSN("edgedb://me:secret@localhost:5656/somedb")
	assert.Equal(t, "secret", opts.Password)
}

func TestMissingPort(t *testing.T) {
	opts := DSN("edgedb://me@localhost/somedb")
	assert.Equal(t, 5656, opts.Port)
}

func TestDialHost(t *testing.T) {
	opts := Options{Host: "some.com", Port: 1234}
	assert.Equal(t, "some.com:1234", opts.dialHost())

	opts = Options{Port: 1234}
	assert.Equal(t, "localhost:1234", opts.dialHost())

	opts = Options{Host: "some.com"}
	assert.Equal(t, "some.com:5656", opts.dialHost())

	opts = Options{}
	assert.Equal(t, "localhost:5656", opts.dialHost())
}
