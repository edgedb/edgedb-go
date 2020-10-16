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
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// conn is initialized by TestMain
var conn *Conn

type serverData struct {
	Port int    `json:"port"`
	Host string `json:"runstate_dir"`
}

var server serverData

func executeOrPanic(command string) {
	err := conn.Execute(command)
	if err != nil {
		panic(err)
	}
}

func startServer() (err error) {
	cmdName := "edgedb-server"
	if slot, ok := os.LookupEnv("EDGEDB_SLOT"); ok {
		cmdName = fmt.Sprintf("%v-%v", cmdName, slot)
	}

	cmd := exec.Command(
		cmdName,
		"--temp-dir",
		"--testmode",
		"--echo-runtime-info",
		"--port=auto",
		"--auto-shutdown",
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	var text string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		text = scanner.Text()
		log.Println(text)
		if strings.HasPrefix(text, "EDGEDB_SERVER_DATA:") {
			break
		}
	}

	encoded := strings.TrimPrefix(text, "EDGEDB_SERVER_DATA:")
	err = json.Unmarshal([]byte(encoded), &server)
	if err != nil {
		if e := cmd.Process.Kill(); e != nil {
			err = fmt.Errorf("%v AND %v", err, e)
		}
		log.Fatal(err)
	}

	return nil
}

func TestMain(m *testing.M) {
	code := 1
	defer func() {
		os.Exit(code)
	}()

	err := startServer()
	if err != nil {
		panic(err)
	}

	conn, err = Connect(Options{
		Host:     server.Host,
		Port:     server.Port,
		User:     "edgedb",
		Database: "edgedb",
		admin:    true,
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	executeOrPanic(`
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
	_ = conn.Execute(`
		CREATE SUPERUSER ROLE user_with_password {
			SET password := 'secret';
		};
	`)
	executeOrPanic("CONFIGURE SYSTEM RESET Auth;")
	executeOrPanic(`
		CONFIGURE SYSTEM INSERT Auth {
			comment := "no password",
			priority := 1,
			method := (INSERT Trust),
			user := {'*'},
		};
	`)
	executeOrPanic(`
		CONFIGURE SYSTEM INSERT Auth {
			comment := "password required",
			priority := 0,
			method := (INSERT SCRAM),
			user := {'user_with_password'}
		}
	`)

	code = m.Run()
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
	assert.Equal(t, [][]int64{{5, 8}}, result)
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
	assert.Equal(t, [][]int64{{5, 8}}, result)
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
	assert.Equal(
		t,
		"[{\"a\" : 0, \"b\" : 1}, {\"a\" : 42, \"b\" : 2}]",
		actual,
	)
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
	expected := &Error{
		Severity: 120,
		Code:     67174656,
		Message:  "Unexpected 'malformed'",
	}
	assert.Equal(t, expected, err)
}
