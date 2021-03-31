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

	"github.com/stretchr/testify/require"
)

func TestRecconnectingConnBorrow(t *testing.T) {
	b := reconnectingConn{}
	err := b.assertUnborrowed()
	require.NoError(t, err)

	_, err = b.borrow("transaction")
	require.NoError(t, err)

	_, err = b.borrow("something else")
	expected := "edgedb.InterfaceError: " +
		"The connection is borrowed for a transaction. " +
		"Use the methods on the transaction object instead."
	require.EqualError(t, err, expected)

	err = b.assertUnborrowed()
	expected = "edgedb.InterfaceError: " +
		"The connection is borrowed for a transaction. " +
		"Use the methods on the transaction object instead."
	require.EqualError(t, err, expected)

	b.unborrow()
	err = b.assertUnborrowed()
	require.NoError(t, err)
}
