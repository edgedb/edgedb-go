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
	"testing"

	types "github.com/geldata/gel-go/internal/geltypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsRead(t *testing.T) {
	creds, err := readCredentials("credentials1.json")
	require.NoError(t, err)

	expected := &credentials{
		database: types.NewOptionalStr("test3n"),
		password: types.NewOptionalStr("lZTBy1RVCfOpBAOwSCwIyBIR"),
		port:     types.NewOptionalInt32(10702),
		user:     "test3n",
	}

	assert.Equal(t, expected, creds)
}

func TestCredentialsEmpty(t *testing.T) {
	creds, err := validateCredentials(map[string]interface{}{})
	assert.EqualError(t, err, "`user` key is required")
	assert.Nil(t, creds)
}

func TestCredentialsPort(t *testing.T) {
	creds, err := validateCredentials(map[string]interface{}{
		"user": "u1",
		"port": "1234",
	})
	assert.EqualError(t, err, "invalid `port` value")
	assert.Nil(t, creds)

	creds, err = validateCredentials(map[string]interface{}{
		"user": "u1",
		"port": 0,
	})
	assert.EqualError(t, err, "invalid `port` value")
	assert.Nil(t, creds)

	creds, err = validateCredentials(map[string]interface{}{
		"user": "u1",
		"port": -1,
	})
	assert.EqualError(t, err, "invalid `port` value")
	assert.Nil(t, creds)

	creds, err = validateCredentials(map[string]interface{}{
		"user": "u1",
		"port": 65536,
	})
	assert.EqualError(t, err, "invalid `port` value")
	assert.Nil(t, creds)
}
