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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryCachingIncludesOutType(t *testing.T) {
	ctx := context.Background()

	err := client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		var result struct {
			Val OptionalTuple `edgedb:"val"`
		}
		e := tx.Execute(ctx, `
			CREATE TYPE Sample {
				CREATE PROPERTY val -> tuple<int64, int64>;
			};
		`)
		assert.NoError(t, e)

		// Run a query with a particular out type
		// that can later be run with a different out type.
		return tx.QuerySingle(ctx, `SELECT Sample { val } LIMIT 1`, &result)
	})
	assert.EqualError(t, err, "edgedb.NoDataError: zero results")

	err = client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		var result struct {
			Val OptionalNamedTuple `edgedb:"val"`
		}

		e := tx.Execute(ctx, `
			CREATE TYPE Sample {
				CREATE PROPERTY val -> tuple<a: int64, b: int64>;
			};
		`)
		assert.NoError(t, e)

		// Run the same query string again with a different out type.
		// There should not be any errors complaining about the out type.
		return tx.QuerySingle(ctx, `SELECT Sample { val } LIMIT 1`, &result)
	})
	assert.EqualError(t, err, "edgedb.NoDataError: zero results")
}

func TestObjectWithoutID(t *testing.T) {
	ctx := context.Background()

	type Database struct {
		Name string `edgedb:"name"`
	}

	var result Database
	err := client.QuerySingle(
		ctx, `
		SELECT sys::Database{ name }
		FILTER .name = 'edgedb'
		LIMIT 1`,
		&result,
	)
	assert.NoError(t, err)
	assert.Equal(t, "edgedb", result.Name)
}

func TestWrongNumberOfArguments(t *testing.T) {
	var result string
	ctx := context.Background()
	err := client.QuerySingle(ctx, `SELECT <str>$0`, &result)
	assert.EqualError(t, err,
		"edgedb.InvalidArgumentError: expected 1 arguments got 0")
}

func TestConnRejectsTransactions(t *testing.T) {
	expected := "edgedb.DisabledCapabilityError: " +
		"cannot execute transaction control commands"

	ctx := context.Background()
	err := client.Execute(ctx, "START TRANSACTION")
	assert.EqualError(t, err, expected)

	var result []byte
	err = client.Query(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = client.QueryJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = client.QuerySingle(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)

	err = client.QuerySingleJSON(ctx, "START TRANSACTION", &result)
	assert.EqualError(t, err, expected)
}

func TestMissmatchedCardinality(t *testing.T) {
	ctx := context.Background()

	var result []int64
	err := client.QuerySingle(ctx, "SELECT {1, 2, 3}", &result)

	expected := "edgedb.ResultCardinalityMismatchError: " +
		"the query has cardinality MANY " +
		"which does not match the expected cardinality ONE"
	assert.EqualError(t, err, expected)
}

func TestMissmatchedResultType(t *testing.T) {
	type C struct { // nolint:unused
		z int // nolint:structcheck
	}

	type B struct { // nolint:unused
		y C // nolint:structcheck
	}

	type A struct {
		x B // nolint:structcheck,unused
	}

	var result A

	ctx := context.Background()
	err := client.QuerySingle(ctx, "SELECT (x := (y := (z := 7)))", &result)

	expected := "edgedb.InvalidArgumentError: " +
		"the \"out\" argument does not match query schema: " +
		"expected edgedb.A.x.y.z to be int64 or edgedb.OptionalInt64 got int"
	assert.EqualError(t, err, expected)
}

// The client should read all messages through ReadyForCommand
// before returning from a QueryX()
func TestParseAllMessagesAfterError(t *testing.T) {
	t.Skip()
	ctx := context.Background()

	// cause error during prepare
	var number float64
	err := client.QuerySingle(ctx, "SELECT 1 / <str>$0", &number, int64(5))

	// nolint:lll
	expected := `edgedb.InvalidTypeError: operator '/' cannot be applied to operands of type 'std::int64' and 'std::str'
query:1:8

SELECT 1 / <str>$0
       ^ Consider using an explicit type cast or a conversion function.`
	assert.EqualError(t, err, expected)

	// cause erroy during execute
	err = client.QuerySingle(ctx, "SELECT 1 / 0", &number)
	assert.EqualError(t, err, "edgedb.DivisionByZeroError: division by zero")

	// cache query so that it is run optimistically next time
	err = client.QuerySingle(ctx, "SELECT 1 / <int64>$0", &number, int64(3))
	assert.NoError(t, err)

	// cause error during optimistic execute
	err = client.QuerySingle(ctx, "SELECT 1 / <int64>$0", &number, int64(0))
	assert.EqualError(t, err, "edgedb.DivisionByZeroError: division by zero")
}

func TestArgumentTypeMissmatch(t *testing.T) {
	type Tuple struct {
		first int16 `edgedb:"0"` // nolint:unused,structcheck
	}

	var res Tuple
	ctx := context.Background()
	err := client.QuerySingle(ctx,
		"SELECT (<int16>$0 + <int16>$1,)", &res, 1, 1111)

	require.NotNil(t, err)
	assert.EqualError(t, err,
		"edgedb.InvalidArgumentError: expected args[0] to be int16, "+
			"edgedb.OptionalInt16 or Int16Marshaler got int")
}

func TestNamedQueryArguments(t *testing.T) {
	ctx := context.Background()
	var result [][]int64
	err := client.Query(
		ctx,
		"SELECT [<int64>$first, <int64>$second]",
		&result,
		map[string]interface{}{
			"first":  int64(5),
			"second": int64(8),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, [][]int64{{5, 8}}, result)
}

func TestNumberedQueryArguments(t *testing.T) {
	ctx := context.Background()
	result := [][]int64{}
	err := client.Query(
		ctx,
		"SELECT [<int64>$0, <int64>$1]",
		&result,
		int64(5),
		int64(8),
	)

	assert.NoError(t, err)
	assert.Equal(t, [][]int64{{5, 8}}, result)
}

func TestQueryJSON(t *testing.T) {
	ctx := context.Background()
	var result []byte
	err := client.QueryJSON(
		ctx,
		"SELECT {(a := 0, b := <int64>$0), (a := 42, b := <int64>$1)}",
		&result,
		int64(1),
		int64(2),
	)

	// casting to string makes error message more helpful
	// when this test fails
	actual := string(result)

	require.NoError(t, err)
	assert.Equal(
		t,
		"[{\"a\" : 0, \"b\" : 1}, {\"a\" : 42, \"b\" : 2}]",
		actual,
	)
}

func TestQuerySingleJSON(t *testing.T) {
	ctx := context.Background()
	var result []byte
	err := client.QuerySingleJSON(
		ctx,
		"SELECT (a := 0, b := <int64>$0)",
		&result,
		int64(42),
	)

	// casting to string makes error messages more helpful
	// when this test fails
	actual := string(result)

	assert.NoError(t, err)
	assert.Equal(t, "{\"a\" : 0, \"b\" : 42}", actual)
}

func TestQuerySingleJSONZeroResults(t *testing.T) {
	ctx := context.Background()
	var result []byte
	err := client.QuerySingleJSON(ctx, "SELECT <int64>{}", &result)

	require.Equal(t, err, errZeroResults)
	assert.Equal(t, []byte(nil), result)
}

func TestQuerySingle(t *testing.T) {
	ctx := context.Background()
	var result int64
	err := client.QuerySingle(ctx, "SELECT 42", &result)

	assert.NoError(t, err)
	assert.Equal(t, int64(42), result)
}

func TestQuerySingleZeroResults(t *testing.T) {
	ctx := context.Background()
	var result int64
	err := client.QuerySingle(ctx, "SELECT <int64>{}", &result)

	assert.Equal(t, errZeroResults, err)
}

func TestQuerySingleNestedSlice(t *testing.T) {
	ctx := context.Background()
	type IDField struct {
		ID UUID `edgedb:"id"`
	}
	type NameField struct {
		Name OptionalStr `edgedb:"name"`
	}
	type UserModel struct {
		IDField   `edgedb:"$inline"`
		NameField `edgedb:"$inline"`
	}
	type UsersField struct {
		Users []UserModel `edgedb:"users"`
	}
	type ViewModel struct {
		UsersField `edgedb:"$inline"`
	}
	result := ViewModel{}
	err := client.QuerySingle(
		ctx,
		`
with a := (INSERT User { name := 'a' }), b := (INSERT User { name := 'b' })
SELECT { users := (SELECT { a, b } { id, name }) }`,
		&result,
	)
	require.NoError(t, err)

	require.Equal(
		t, 2, len(result.Users),
		"wrong number of users, expected 2 got %v", len(result.Users))
	assert.NotEqual(t, result.Users[0].ID, UUID{})
	a, _ := result.Users[0].Name.Get()
	assert.Equal(t, a, "a")

	assert.NotEqual(t, result.Users[1].ID, UUID{})
	assert.NotEqual(t, result.Users[0].ID, result.Users[1].ID)
	b, _ := result.Users[1].Name.Get()
	assert.Equal(t, b, "b")
}

func TestError(t *testing.T) {
	ctx := context.Background()
	err := client.Execute(ctx, "malformed query;")
	assert.EqualError(
		t,
		err,
		`edgedb.EdgeQLSyntaxError: Unexpected 'malformed'
query:1:1

malformed query;
^ error`,
	)

	var expected Error
	assert.True(t, errors.As(err, &expected))
}

func TestQueryTimesOut(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()

	var r int64
	err := client.QuerySingle(ctx, "SELECT 1;", &r)
	require.True(
		t,
		errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(err, os.ErrDeadlineExceeded),
		err,
	)
	require.Equal(t, int64(0), r)

	err = client.QuerySingle(context.Background(), "SELECT 2;", &r)
	require.NoError(t, err)
	assert.Equal(t, int64(2), r)
}

func TestNilResultValue(t *testing.T) {
	ctx := context.Background()
	err := client.Query(ctx, "SELECT 1", nil)
	assert.EqualError(t, err, "edgedb.InterfaceError: "+
		"the \"out\" argument must be a pointer, got untyped nil")
}

func TestExecutWithArgs(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	err := client.Execute(ctx, "select <int64>$0; select <int64>$0;", int64(1))
	assert.NoError(t, err)

	err = client.Tx(ctx, func(ctx context.Context, tx *Tx) error {
		err = tx.Execute(ctx, "select <int64>$0; select <int64>$0;", int64(1))
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)
}

func TestClientRejectsSessionConfig(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	expected := "edgedb.DisabledCapabilityError: " +
		"cannot execute session configuration queries"

	ctx := context.Background()
	err := client.Execute(ctx, "SET ALIAS bar AS MODULE std")
	assert.EqualError(t, err, expected)

	var result []byte
	err = client.Query(ctx, "SET ALIAS bar AS MODULE std", &result)
	assert.EqualError(t, err, expected)

	err = client.QueryJSON(ctx, "SET ALIAS bar AS MODULE std", &result)
	assert.EqualError(t, err, expected)

	err = client.QuerySingle(ctx, "SET ALIAS bar AS MODULE std", &result)
	assert.EqualError(t, err, expected)

	err = client.QuerySingleJSON(ctx, "SET ALIAS bar AS MODULE std", &result)
	assert.EqualError(t, err, expected)
}

func TestWithConfig(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	query := "SELECT assert_single(cfg::Config.query_execution_timeout)"

	var result Duration
	err := client.QuerySingle(ctx, query, &result)
	require.NoError(t, err)
	assert.Equal(t, Duration(0), result)

	a := client.WithConfig(map[string]interface{}{
		"query_execution_timeout": Duration(65_432_000),
	})
	err = a.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(65_432_000), result)

	err = client.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(0), result)

	b := a.WithConfig(map[string]interface{}{
		"query_execution_timeout": Duration(32_100_000),
	})
	err = b.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(32_100_000), result)

	err = a.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(65_432_000), result)

	err = client.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(0), result)
}

func TestInvalidWithConfig(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	var result int64

	err := client.QuerySingle(ctx, "select 1", &result)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)

	c := client.WithConfig(map[string]interface{}{"hello": "world"})
	err = c.QuerySingle(ctx, "select 1", &result)
	assert.EqualError(t, err, "edgedb.BinaryProtocolError: "+
		"invalid connection state: "+
		"found unknown state value state.config.hello")

	err = client.QuerySingle(ctx, "select 1", &result)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)

	c = client.WithConfig(map[string]interface{}{
		"query_execution_timeout": "this should be Duration not string",
	})
	err = c.QuerySingle(ctx, "select 1", &result)
	assert.EqualError(t, err, "edgedb.BinaryProtocolError: "+
		"invalid connection state: "+
		"expected state.config.query_execution_timeout to be edgedb.Duration "+
		"got: string")

	err = client.QuerySingle(ctx, "select 1", &result)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)
}

func TestWithoutConfig(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	query := "SELECT assert_single(cfg::Config.query_execution_timeout)"

	var result Duration
	err := client.QuerySingle(ctx, query, &result)
	require.NoError(t, err)
	assert.Equal(t, Duration(0), result)

	a := client.WithConfig(map[string]interface{}{
		"query_execution_timeout": Duration(65_432_000),
	})
	err = a.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(65_432_000), result)

	b := a.WithoutConfig("query_execution_timeout")
	err = b.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(0), result)

	err = a.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(65_432_000), result)

	err = client.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(0), result)

	b = client.WithoutConfig("some", "crazy", "names")
	err = b.QuerySingle(ctx, query, &result)
	assert.NoError(t, err)
	assert.Equal(t, Duration(0), result)
}

func TestWithModuleAliases(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	var result int64

	err := client.QuerySingle(
		ctx,
		"SELECT <my_new_name_for_std::int64>1",
		&result,
	)
	assert.EqualError(t, err, "edgedb.InvalidReferenceError: "+
		"type 'my_new_name_for_std::int64' does not exist\n"+
		"query:1:9\n\n"+
		"SELECT <my_new_name_for_std::int64>1\n"+
		"        ^ error")

	a := client.WithModuleAliases(ModuleAlias{"my_new_name_for_std", "std"})

	err = a.QuerySingle(ctx, "SELECT <my_new_name_for_std::int64>2", &result)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result)

	err = client.QuerySingle(
		ctx,
		"SELECT <my_new_name_for_std::int64>3",
		&result,
	)
	assert.EqualError(t, err, "edgedb.InvalidReferenceError: "+
		"type 'my_new_name_for_std::int64' does not exist\n"+
		"query:1:9\n\n"+
		"SELECT <my_new_name_for_std::int64>3\n"+
		"        ^ error")

	b := a.WithModuleAliases(ModuleAlias{"my_new_name_for_std", "math"})
	err = b.QuerySingle(ctx, "SELECT <my_new_name_for_std::int64>4", &result)
	assert.EqualError(t, err, "edgedb.InvalidReferenceError: "+
		"type 'my_new_name_for_std::int64' does not exist\n"+
		"query:1:9\n\n"+
		"SELECT <my_new_name_for_std::int64>4\n"+
		"        ^ error")

	err = a.QuerySingle(ctx, "SELECT <my_new_name_for_std::int64>5", &result)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), result)

	err = client.QuerySingle(
		ctx,
		"SELECT <my_new_name_for_std::int64>6",
		&result,
	)
	assert.EqualError(t, err, "edgedb.InvalidReferenceError: "+
		"type 'my_new_name_for_std::int64' does not exist\n"+
		"query:1:9\n\n"+
		"SELECT <my_new_name_for_std::int64>6\n"+
		"        ^ error")
}

func TestInvalidWithModuleAliases(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	var result int64

	a := client.WithModuleAliases(ModuleAlias{
		"my_alias", "this_module_doesnt_exist"})

	err := a.QuerySingle(ctx, "SELECT <my_alias::int64>1", &result)
	assert.EqualError(t, err, "edgedb.InvalidReferenceError: "+
		"type 'my_alias::int64' does not exist\n"+
		"query:1:9\n\n"+
		"SELECT <my_alias::int64>1\n"+
		"        ^ error")
}

func TestWithoutModuleAliases(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	var result int64

	a := client.WithModuleAliases(ModuleAlias{"my_new_name_for_std", "std"})
	b := a.WithoutModuleAliases("my_new_name_for_std")

	err := b.QuerySingle(ctx, "SELECT <my_new_name_for_std::int64>4", &result)
	assert.EqualError(t, err, "edgedb.InvalidReferenceError: "+
		"type 'my_new_name_for_std::int64' does not exist\n"+
		"query:1:9\n\n"+
		"SELECT <my_new_name_for_std::int64>4\n"+
		"        ^ error")

	err = a.QuerySingle(ctx, "SELECT <my_new_name_for_std::int64>5", &result)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), result)

	err = client.QuerySingle(
		ctx,
		"SELECT <my_new_name_for_std::int64>6",
		&result,
	)
	assert.EqualError(t, err, "edgedb.InvalidReferenceError: "+
		"type 'my_new_name_for_std::int64' does not exist\n"+
		"query:1:9\n\n"+
		"SELECT <my_new_name_for_std::int64>6\n"+
		"        ^ error")
}

func TestWithGlobals(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	var result string

	err := client.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "default", result)

	a := client.WithGlobals(map[string]interface{}{
		"default::global_value": "first",
	})
	err = a.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "first", result)

	err = client.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "default", result)

	b := a.WithGlobals(map[string]interface{}{
		"default::global_value": "second",
	})
	err = b.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "second", result)

	err = a.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "first", result)

	err = client.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "default", result)
}

func TestInvalidWithGlobals(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	var result int64

	a := client.WithGlobals(map[string]interface{}{
		"default::this": "thing donesnt exist",
	})

	err := a.QuerySingle(ctx, "SELECT GLOBAL this", &result)
	assert.EqualError(t, err, "edgedb.BinaryProtocolError: "+
		"invalid connection state: "+
		"found unknown state value state.globals.default::this")

	b := client.WithGlobals(map[string]interface{}{
		"default::global_value": 27,
	})

	err = b.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	assert.EqualError(t, err, "edgedb.BinaryProtocolError: "+
		"invalid connection state: "+
		"expected state.globals.default::global_value to be string "+
		"got: int")
}

func TestWithoutGlobals(t *testing.T) {
	if protocolVersion.LT(protocolVersion1p0) {
		t.Skip()
	}

	ctx := context.Background()
	var result string

	err := client.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "default", result)

	a := client.WithGlobals(map[string]interface{}{
		"default::global_value": "first",
	})

	b := a.WithoutGlobals("default::global_value")
	err = b.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "default", result)

	err = a.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "first", result)

	err = client.QuerySingle(ctx, "SELECT GLOBAL global_value", &result)
	require.NoError(t, err)
	assert.Equal(t, "default", result)
}

func TestWithConfigWrongServerVersion(t *testing.T) {
	if protocolVersion.GTE(protocolVersion1p0) {
		t.Skip()
	}
	ctx := context.Background()
	var result int64

	a := client.WithGlobals(map[string]interface{}{
		"default::global_value": "first",
	})
	err := a.QuerySingle(ctx, "SELECT 1", &result)
	require.EqualError(t, err, "edgedb.InterfaceError: "+
		"client methods WithConfig, WithGlobals, and WithModuleAliases "+
		"are not supported by the server. "+
		"Upgrade your server to version 2.0 or greater to use these features.")

	b := client.WithModuleAliases(ModuleAlias{"other_math", "math"})
	err = b.QuerySingle(ctx, "SELECT 1", &result)
	require.EqualError(t, err, "edgedb.InterfaceError: "+
		"client methods WithConfig, WithGlobals, and WithModuleAliases "+
		"are not supported by the server. "+
		"Upgrade your server to version 2.0 or greater to use these features.")

	c := client.WithConfig(map[string]interface{}{
		"query_execution_timeout": Duration(65_432_000),
	})
	err = c.QuerySingle(ctx, "SELECT 1", &result)
	require.EqualError(t, err, "edgedb.InterfaceError: "+
		"client methods WithConfig, WithGlobals, and WithModuleAliases "+
		"are not supported by the server. "+
		"Upgrade your server to version 2.0 or greater to use these features.")
}
