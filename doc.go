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

// Package edgedb is the official Go driver for [EdgeDB].
//
// Typical usage looks like this:
//
//	package main
//
//	import (
//	    "context"
//	    "log"
//
//	    "github.com/edgedb/edgedb-go"
//	)
//
//	func main() {
//	    ctx := context.Background()
//	    client, err := edgedb.CreateClient(ctx, edgedb.Options{})
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer client.Close()
//
//	    var (
//	        age   int64 = 21
//	        users []struct {
//	            ID   edgedb.UUID `edgedb:"id"`
//	            Name string      `edgedb:"name"`
//	        }
//	    )
//
//	    query := "SELECT User{name} FILTER .age = <int64>$0"
//	    err = client.Query(ctx, query, &users, age)
//	    ...
//	}
//
// We recommend using environment variables for connection parameters. See the
// [client connection docs] for more information.
//
// You may also connect to a database using a DSN:
//
//	url := "edgedb://edgedb@localhost/edgedb"
//	client, err := edgedb.CreateClientDSN(ctx, url, opts)
//
// Or you can use Option fields.
//
//	opts := edgedb.Options{
//	    Database:    "edgedb",
//	    User:        "edgedb",
//	    Concurrency: 4,
//	}
//
//	client, err := edgedb.CreateClient(ctx, opts)
//
// # Errors
//
// edgedb never returns underlying errors directly.
// If you are checking for things like context expiration
// use errors.Is() or errors.As().
//
//	err := client.Query(...)
//	if errors.Is(err, context.Canceled) { ... }
//
// Most errors returned by the edgedb package will satisfy the edgedb.Error
// interface which has methods for introspecting.
//
//	err := client.Query(...)
//
//	var edbErr edgedb.Error
//	if errors.As(err, &edbErr) && edbErr.Category(edgedb.NoDataError){
//	    ...
//	}
//
// # Datatypes
//
// The following list shows the marshal/unmarshal
// mapping between EdgeDB types and go types:
//
//	EdgeDB                   Go
//	---------                ---------
//	Set                      []anytype
//	array<anytype>           []anytype
//	tuple                    struct
//	named tuple              struct
//	Object                   struct
//	bool                     bool, edgedb.OptionalBool
//	bytes                    []byte, edgedb.OptionalBytes
//	str                      string, edgedb.OptionalStr
//	anyenum                  string, edgedb.OptionalStr
//	datetime                 time.Time, edgedb.OptionalDateTime
//	cal::local_datetime      edgedb.LocalDateTime,
//	                         edgedb.OptionalLocalDateTime
//	cal::local_date          edgedb.LocalDate, edgedb.OptionalLocalDate
//	cal::local_time          edgedb.LocalTime, edgedb.OptionalLocalTime
//	duration                 edgedb.Duration, edgedb.OptionalDuration
//	cal::relative_duration   edgedb.RelativeDuration,
//	                         edgedb.OptionalRelativeDuration
//	float32                  float32, edgedb.OptionalFloat32
//	float64                  float64, edgedb.OptionalFloat64
//	int16                    int16, edgedb.OptionalFloat16
//	int32                    int32, edgedb.OptionalInt16
//	int64                    int64, edgedb.OptionalInt64
//	uuid                     edgedb.UUID, edgedb.OptionalUUID
//	json                     []byte, edgedb.OptionalBytes
//	bigint                   *big.Int, edgedb.OptionalBigInt
//
//	decimal                  user defined (see Custom Marshalers)
//
// Note that EdgeDB's std::duration type is represented in int64 microseconds
// while go's time.Duration type is int64 nanoseconds. It is incorrect to cast
// one directly to the other.
//
// Shape fields that are not required must use optional types for receiving
// query results. The edgedb.Optional struct can be embedded to make structs
// optional.
//
//	type User Struct {
//	    edgedb.Optional
//	    Email string `edgedb:"email"`
//	}
//
//	var result User
//	err := client.QuerySingle(ctx, `SELECT User { email } LIMIT 0`, $result)
//	fmt.Println(result.Missing())
//	// Output: true
//
//	err := client.QuerySingle(ctx, `SELECT User { email } LIMIT 1`, $result)
//	fmt.Println(result.Missing())
//	// Output: false
//
// Not all types listed above are valid query parameters.  To pass a slice of
// scalar values use array in your query. EdgeDB doesn't currently support
// using sets as parameters.
//
//	query := `select User filter .id in array_unpack(<array<uuid>>$1)`
//	client.QuerySingle(ctx, query, $user, []edgedb.UUID{...})
//
// Nested structures are also not directly allowed but you can use [json]
// instead.
//
// # Custom Marshalers
//
// Interfaces for user defined marshaler/unmarshalers  are documented in the
// internal/marshal package.
//
// [EdgeDB]: https://www.edgedb.com
// [json]: https://www.edgedb.com/docs/edgeql/insert#bulk-inserts
// [client connection docs]: https://www.edgedb.com/docs/clients/connection
package edgedb
