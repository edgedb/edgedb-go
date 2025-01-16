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

// Package gel is the official Go driver for [Gel]. Additionally,
// [github.com/geldata/gel-go/cmd/edgeql-go] is a code generator that
// generates go functions from edgeql files.
//
// Typical client usage looks like this:
//
//	package main
//
//	import (
//	    "context"
//	    "log"
//
//	    "github.com/geldata/gel-go"
//	)
//
//	func main() {
//	    ctx := context.Background()
//	    client, err := gel.CreateClient(ctx, gel.Options{})
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer client.Close()
//
//	    var (
//	        age   int64 = 21
//	        users []struct {
//	            ID   gel.UUID `gel:"id"`
//	            Name string   `gel:"name"`
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
//	url := "gel://edgedb@localhost/edgedb"
//	client, err := gel.CreateClientDSN(ctx, url, opts)
//
// Or you can use Option fields.
//
//	opts := gel.Options{
//	    Database:    "edgedb",
//	    User:        "edgedb",
//	    Concurrency: 4,
//	}
//
//	client, err := gel.CreateClient(ctx, opts)
//
// # Errors
//
// gel never returns underlying errors directly.
// If you are checking for things like context expiration
// use errors.Is() or errors.As().
//
//	err := client.Query(...)
//	if errors.Is(err, context.Canceled) { ... }
//
// Most errors returned by the gel package will satisfy the gel.Error
// interface which has methods for introspecting.
//
//	err := client.Query(...)
//
//	var gelErr gel.Error
//	if errors.As(err, &gelErr) && gelErr.Category(gel.NoDataError){
//	    ...
//	}
//
// # Datatypes
//
// The following list shows the marshal/unmarshal
// mapping between Gel types and go types:
//
//	Gel                      Go
//	---------                ---------
//	Set                      []anytype
//	array<anytype>           []anytype
//	tuple                    struct
//	named tuple              struct
//	Object                   struct
//	bool                     bool, gel.OptionalBool
//	bytes                    []byte, gel.OptionalBytes
//	str                      string, gel.OptionalStr
//	anyenum                  string, gel.OptionalStr
//	datetime                 time.Time, gel.OptionalDateTime
//	cal::local_datetime      gel.LocalDateTime,
//	                         gel.OptionalLocalDateTime
//	cal::local_date          gel.LocalDate, gel.OptionalLocalDate
//	cal::local_time          gel.LocalTime, gel.OptionalLocalTime
//	duration                 gel.Duration, gel.OptionalDuration
//	cal::relative_duration   gel.RelativeDuration,
//	                         gel.OptionalRelativeDuration
//	float32                  float32, gel.OptionalFloat32
//	float64                  float64, gel.OptionalFloat64
//	int16                    int16, gel.OptionalFloat16
//	int32                    int32, gel.OptionalInt16
//	int64                    int64, gel.OptionalInt64
//	uuid                     gel.UUID, gel.OptionalUUID
//	json                     []byte, gel.OptionalBytes
//	bigint                   *big.Int, gel.OptionalBigInt
//
//	decimal                  user defined (see Custom Marshalers)
//
// Note that Gel's std::duration type is represented in int64 microseconds
// while go's time.Duration type is int64 nanoseconds. It is incorrect to cast
// one directly to the other.
//
// Shape fields that are not required must use optional types for receiving
// query results. The gel.Optional struct can be embedded to make structs
// optional.
//
//	type User struct {
//	    gel.Optional
//	    Email string `gel:"email"`
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
// scalar values use array in your query. Gel doesn't currently support
// using sets as parameters.
//
//	query := `select User filter .id in array_unpack(<array<uuid>>$1)`
//	client.QuerySingle(ctx, query, $user, []gel.UUID{...})
//
// Nested structures are also not directly allowed but you can use [json]
// instead.
//
// By default Gel will ignore embedded structs when marshaling/unmarshaling.
// To treat an embedded struct's fields as part of the parent struct's fields,
// tag the embedded struct with `gel:"$inline"`.
//
//	type Object struct {
//	    ID gel.UUID
//	}
//
//	type User struct {
//	    Object `gel:"$inline"`
//	    Name string
//	}
//
// # Custom Marshalers
//
// Interfaces for user defined marshaler/unmarshalers  are documented in the
// internal/marshal package.
//
// [Gel]: https://www.edgedb.com
// [json]: https://www.edgedb.com/docs/edgeql/insert#bulk-inserts
// [client connection docs]: https://www.edgedb.com/docs/clients/connection
package gel
