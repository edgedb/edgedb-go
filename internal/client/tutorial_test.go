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
	"context"
	"fmt"
	"math/rand"
	"testing"

	types "github.com/edgedb/edgedb-go/internal/geltypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Person struct {
	ID        types.UUID `gel:"id"`
	FirstName string     `gel:"first_name"`
	LastName  string     `gel:"last_name"`
}

type Movie struct {
	ID       types.UUID          `gel:"id"`
	Title    string              `gel:"title"`
	Year     types.OptionalInt64 `gel:"year"`
	Director Person              `gel:"director"`
	Actors   []Person            `gel:"actors"`
}

func TestTutorial(t *testing.T) {
	ctx := context.Background()
	dbName := fmt.Sprintf("test%v", rand.Intn(10_000))
	err := client.Execute(ctx, "CREATE DATABASE "+dbName)
	require.NoError(t, err)

	edb, err := CreateClient(
		ctx,
		Options{
			Host:       opts.Host,
			Port:       opts.Port,
			User:       opts.User,
			Password:   opts.Password,
			Database:   dbName,
			TLSOptions: opts.TLSOptions,
		},
	)
	require.NoError(t, err)

	err = edb.Execute(ctx,
		`START MIGRATION TO {
			module default {
				type Movie {
					required property title -> str;
					# the year of release
					property year -> int64;
					required link director -> Person;
					multi link actors -> Person;
				}
				type Person {
					required property first_name -> str;
					required property last_name -> str;
				}
			}
		};
		POPULATE MIGRATION;
		COMMIT MIGRATION;`)
	require.NoError(t, err)

	err = edb.Execute(ctx, `
		INSERT Movie {
			title := 'Blade Runner 2049',
			year := 2017,
			director := (
				INSERT Person {
					first_name := 'Denis',
					last_name := 'Villeneuve',
				}
			),
			actors := {
				(INSERT Person {
					first_name := 'Harrison',
					last_name := 'Ford',
				}),
				(INSERT Person {
					first_name := 'Ryan',
					last_name := 'Gosling',
				}),
				(INSERT Person {
					first_name := 'Ana',
					last_name := 'de Armas',
				}),
			}
		}`,
	)
	require.NoError(t, err)

	err = edb.Execute(ctx, `
		INSERT Movie {
				title := 'Dune',
				director := (
						SELECT Person
						FILTER
								# the last name is sufficient
								# to identify the right person
								.last_name = 'Villeneuve'
						# the LIMIT is needed to satisfy the single
						# link requirement validation
						LIMIT 1
				)
		};`,
	)
	require.NoError(t, err)

	var out []Movie
	err = edb.Query(ctx, `
		SELECT Movie {
				title,
				year,
				director: {
						first_name,
						last_name
				},
				actors: {
						first_name,
						last_name
				}
		}`,
		&out,
	)
	require.NoError(t, err)

	// clobber IDs with zero value since they are not deterministic
	zeroID := make([]byte, 16)
	for i := 0; i < len(out); i++ {
		copy(out[i].ID[:], zeroID)
		copy(out[i].Director.ID[:], zeroID)
		for j := 0; j < len(out[i].Actors); j++ {
			copy(out[i].Actors[j].ID[:], zeroID)
		}
	}

	expected := []Movie{
		{
			Title: "Blade Runner 2049",
			Director: Person{
				FirstName: "Denis",
				LastName:  "Villeneuve",
			},
			Actors: []Person{
				{
					FirstName: "Harrison",
					LastName:  "Ford",
				},
				{
					FirstName: "Ryan",
					LastName:  "Gosling",
				},
				{
					FirstName: "Ana",
					LastName:  "de Armas",
				},
			},
		},
		{
			Title: "Dune",
			Director: Person{
				FirstName: "Denis",
				LastName:  "Villeneuve",
			},
			Actors: []Person{},
		},
	}
	expected[0].Year.Set(2017)

	assert.Equal(t, expected, out)
	assert.NoError(t, edb.Close())
}
