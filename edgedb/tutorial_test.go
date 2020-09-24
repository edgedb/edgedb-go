// queries taken from https://www.edgedb.com/docs/tutorial/createdb/
package main

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/fmoor/edgedb-golang/edgedb/protocol"
	"github.com/stretchr/testify/assert"
)

func TestCreateDB(t *testing.T) {
	options := ConnConfig{"edgedb", "edgedb"}
	edb, err := Connect(options)
	assert.Nil(t, err)
	defer edb.Close()

	rand.Seed(time.Now().UnixNano())
	dbName := fmt.Sprintf("test%v", rand.Intn(10_000))
	result, err := edb.Query("CREATE DATABASE " + dbName + ";")
	assert.Nil(t, err)
	defer func() {
		result, err := edb.Query("DROP DATABASE " + dbName + ";")
		assert.Nil(t, err)
		assert.Equal(t, []interface{}{}, result)
	}()
	assert.Equal(t, []interface{}{}, result)

	options = ConnConfig{dbName, "edgedb"}
	edb2, err := Connect(options)
	assert.Nil(t, err)
	defer edb2.Close()

	withDB := func(fun func(t *testing.T, edb *Conn)) func(t *testing.T) {
		return func(t *testing.T) {
			fun(t, edb2)
		}
	}

	t.Run("migrate", withDB(testMigration))
	t.Run("insert movie", withDB(testInsertMovie))
}

func testMigration(t *testing.T, edb *Conn) {
	result, err := edb.Query(`
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
		};
	`)
	assert.Nil(t, err)
	assert.Equal(t, []interface{}{}, result)

	result, err = edb.Query(`POPULATE MIGRATION;`)
	assert.Nil(t, err)
	assert.Equal(t, []interface{}{}, result)

	result, err = edb.Query(`COMMIT MIGRATION;`)
	assert.Nil(t, err)
	assert.Equal(t, []interface{}{}, result)

	result, err = edb.Query(`
		SELECT schema::ObjectType.name
		FILTER schema::ObjectType.name LIKE 'default::%'
	`)
	assert.Nil(t, err)
	assert.Equal(t, []interface{}{"default::Movie", "default::Person"}, result)
}

func testInsertMovie(t *testing.T, edb *Conn) {
	result, err := edb.Query(`
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
		};
	`)
	assert.Nil(t, err)

	object := result.([]interface{})[0].(map[string]interface{})
	id := object["id"]
	// ids are not deterministic so just check that there is one
	assert.IsType(t, protocol.UUID(""), id)

	delete(object, "id")
	expected := []interface{}{map[string]interface{}{}}
	assert.Equal(t, expected, result)
}
