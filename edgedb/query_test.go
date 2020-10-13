package edgedb

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// conn is initialized by TestMain
var conn *Conn

func executeOrFatal(command string) {
	err := conn.Execute(command)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	opts := DSN("edgedb://edgedb@localhost:5656/edgedb")
	var err error
	conn, err = Connect(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	executeOrFatal(`
		START MIGRATION TO {
			module default {
				type User {
					property name -> str;
				}
			}
		};
		POPULATE MIGRATION;
		COMMIT MIGRATION;
	`)
	conn.Execute(`
		CREATE SUPERUSER ROLE user_with_password {
			SET password := 'secret';
		};
	`)
	executeOrFatal("CONFIGURE SYSTEM RESET Auth;")
	executeOrFatal(`
		CONFIGURE SYSTEM INSERT Auth {
			comment := "no password",
			priority := 1,
			method := (INSERT Trust),
			user := {'*'},
		};
	`)
	executeOrFatal(`
		CONFIGURE SYSTEM INSERT Auth {
			comment := "password required",
			priority := 0,
			method := (INSERT SCRAM),
			user := {'user_with_password'}
		}
	`)

	os.Exit(m.Run())
}

func TestNamedQueryArguments(t *testing.T) {
	result := [][]int64{}
	err := conn.Query(
		"SELECT [<int64>$first, <int64>$second]",
		&result,
		map[string]interface{}{
			"first":  int64(5),
			"second": int64(8),
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, [][]int64{[]int64{5, 8}}, result)
}

func TestNumberedQueryArguments(t *testing.T) {
	result := [][]int64{}
	err := conn.Query(
		"SELECT [<int64>$0, <int64>$1]",
		&result,
		int64(5),
		int64(8),
	)

	assert.Nil(t, err)
	assert.Equal(t, [][]int64{[]int64{5, 8}}, result)
}

func TestQueryJSON(t *testing.T) {
	result, err := conn.QueryJSON(
		"SELECT {(a := 0, b := <int64>$0), (a := 42, b := <int64>$1)}",
		int64(1),
		int64(2),
	)

	// casting to string makes error message more helpful
	// when this test fails
	actual := string(result)

	assert.Nil(t, err)
	assert.Equal(t, "[{\"a\" : 0, \"b\" : 1}, {\"a\" : 42, \"b\" : 2}]", actual)
}

func TestQueryOneJSON(t *testing.T) {
	result, err := conn.QueryOneJSON(
		"SELECT (a := 0, b := <int64>$0)",
		int64(42),
	)

	// casting to string makes error messages more helpful
	// when this test fails
	actual := string(result)

	assert.Nil(t, err)
	assert.Equal(t, "{\"a\" : 0, \"b\" : 42}", actual)
}

func TestQueryOneJSONZeroResults(t *testing.T) {
	result, err := conn.QueryOneJSON("SELECT <int64>{}")

	assert.Equal(t, err, ErrorZeroResults)
	assert.Equal(t, []byte(nil), result)
}

func TestQueryOne(t *testing.T) {
	var result int64
	err := conn.QueryOne("SELECT 42", &result)

	assert.Nil(t, err)
	assert.Equal(t, int64(42), result)
}

func TestQueryOneZeroResults(t *testing.T) {
	result := (*int64)(nil)
	err := conn.QueryOne("SELECT <int64>{}", result)

	assert.Equal(t, ErrorZeroResults, err)
	assert.Nil(t, result)
}

func TestError(t *testing.T) {
	err := conn.Execute("malformed query;")
	expected := &Error{Severity: 120, Code: 67174656, Message: "Unexpected 'malformed'"}
	assert.Equal(t, expected, err)
}
