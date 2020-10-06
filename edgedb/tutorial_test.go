package edgedb

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/fmoor/edgedb-golang/edgedb/options"
	"github.com/fmoor/edgedb-golang/edgedb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Person struct {
	FirstName string `edgedb:"first_name"`
	LastName  string `edgedb:"last_name"`
}

type Movie struct {
	Title    string   `edgedb:"title"`
	Year     int64    `edgedb:"year"`
	Director Person   `edgedb:"director"`
	Actors   []Person `edgedb:"actors"`
}

func TestTutorial(t *testing.T) {
	opts := options.FromDSN("edgedb://edgedb@localhost:5656/edgedb")
	conn0, err := Connect(opts)
	require.Nil(t, err)
	defer conn0.Close()

	rand.Seed(time.Now().UnixNano())
	dbName := fmt.Sprintf("test%v", rand.Intn(10_000))
	err = conn0.Query("CREATE DATABASE "+dbName+";", &[]interface{}{})
	require.Nil(t, err)

	defer func() {
		conn0.Query("DROP DATABASE "+dbName+";", &types.Set{})
	}()

	opts = options.Options{Database: dbName, User: "edgedb"}
	conn, _ := Connect(opts)
	defer conn.Close()

	err = conn.Query(`
		START MIGRATION TO {
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
		};`,
		&[]interface{}{},
	)
	require.Nil(t, err)

	err = conn.Query(`POPULATE MIGRATION;`, &[]interface{}{})
	require.Nil(t, err)

	err = conn.Query(`COMMIT MIGRATION;`, &[]interface{}{})
	require.Nil(t, err)

	err = conn.Query(`
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
		};`,
		&[]interface{}{},
	)
	require.Nil(t, err)

	err = conn.Query(`
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
		&[]interface{}{},
	)
	require.Nil(t, err)

	var out []Movie

	err = conn.Query(`
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
	require.Nil(t, err)

	expected := []Movie{
		Movie{
			Title: "Blade Runner 2049",
			Year:  int64(2017),
			Director: Person{
				FirstName: "Denis",
				LastName:  "Villeneuve",
			},
			Actors: []Person{
				Person{
					FirstName: "Harrison",
					LastName:  "Ford",
				},
				Person{
					FirstName: "Ryan",
					LastName:  "Gosling",
				},
				Person{
					FirstName: "Ana",
					LastName:  "de Armas",
				},
			},
		},
		Movie{
			Title: "Dune",
			Director: Person{
				FirstName: "Denis",
				LastName:  "Villeneuve",
			},
			Actors: []Person{},
		},
	}

	assert.Equal(t, expected, out)
}
