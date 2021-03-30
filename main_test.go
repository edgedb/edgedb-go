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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// initialized by TestMain
var (
	opts Options
	conn *Conn
)

func executeOrPanic(command string) {
	ctx := context.Background()
	err := conn.Execute(ctx, command)
	if err != nil {
		panic(err)
	}
}

func startServer() (err error) {
	log.Print("starting new server")

	cmdName := os.Getenv("EDGEDB_SERVER_BIN")
	if cmdName == "" {
		log.Fatal("EDGEDB_SERVER_BIN not set")
	}

	cmdArgs := []string{
		"--temp-dir",
		"--testmode",
		"--echo-runtime-info",
		"--port=auto",
		"--auto-shutdown",
		`--bootstrap-command=` +
			`CREATE SUPERUSER ROLE test { SET password := "shhh"  }`,
	}

	log.Println(cmdName, strings.Join(cmdArgs, " "))

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stderr = os.Stderr
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
		fmt.Println(text)
		if strings.HasPrefix(text, "EDGEDB_SERVER_DATA:") {
			break
		}
	}

	type serverData struct {
		Port int    `json:"port"`
		Host string `json:"runstate_dir"`
	}

	var data serverData
	encoded := strings.TrimPrefix(text, "EDGEDB_SERVER_DATA:")
	err = json.Unmarshal([]byte(encoded), &data)
	if err != nil {
		if e := cmd.Process.Kill(); e != nil {
			err = fmt.Errorf("%v AND %v", err, e)
		}
		log.Fatal(err)
	}

	opts = Options{
		Hosts:    []string{data.Host},
		Ports:    []int{data.Port},
		User:     "test",
		Password: "shhh",
		Database: "edgedb",
	}

	log.Print("server started")
	return nil
}

func TestMain(m *testing.M) {
	var err error = nil
	code := 1
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			fmt.Println(debug.Stack())
		}

		if err != nil {
			log.Println("error while cleaning up: ", err)
		}
		os.Exit(code)
	}()

	err = startServer()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	log.Println("connecting")
	conn, err = ConnectOne(ctx, opts)
	if err != nil {
		panic(err)
	}
	log.Println("connected")

	defer conn.Close() // nolint:errcheck

	log.Println("running migration")
	executeOrPanic(`
			START MIGRATION TO {
				module default {
					type User {
						property name -> str;
					}
					type TxTest {
						required property name -> str;
					}
					scalar type CustomInt64 extending int64;
					scalar type ColorEnum extending enum<Red, Green, Blue>;
				}
			};
			POPULATE MIGRATION;
			COMMIT MIGRATION;
		`)
	executeOrPanic(`
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

	rand.Seed(time.Now().Unix())
	log.Println("starting tests")
	code = m.Run()
}
