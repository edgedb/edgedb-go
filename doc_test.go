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

package gel_test

import (
	"context"
	"fmt"
	"log"
	"time"

	gel "github.com/geldata/gel-go"
)

type User struct {
	ID   gel.UUID  `gel:"id"`
	Name string    `gel:"name"`
	DOB  time.Time `gel:"dob"`
}

func Example() {
	opts := gel.Options{Concurrency: 4}
	ctx := context.Background()
	db, err := gel.CreateClientDSN(ctx, "gel://edgedb@localhost/test", opts)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create a user object type.
	err = db.Execute(ctx, `
		CREATE TYPE User {
			CREATE REQUIRED PROPERTY name -> str;
			CREATE PROPERTY dob -> datetime;
		}
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Insert a new user.
	var inserted struct{ id gel.UUID }
	err = db.QuerySingle(ctx, `
		INSERT User {
			name := <str>$0,
			dob := <datetime>$1
		}
	`, &inserted, "Bob", time.Date(1984, 3, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		log.Fatal(err)
	}

	// Select users.
	var users []User
	args := map[string]interface{}{"name": "Bob"}
	query := "SELECT User {name, dob} FILTER .name = <str>$name"
	err = db.Query(ctx, query, &users, args)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(users)
}
