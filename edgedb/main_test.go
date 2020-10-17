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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// initialized by TestMain
var (
	server Options
	conn   *Conn
)

func executeOrPanic(command string) {
	err := conn.Execute(command)
	if err != nil {
		panic(err)
	}
}

func getLocalServer() error {
	credFileName, ok := os.LookupEnv("EDGEDB_CREDENTIALS_FILE")
	if !ok {
		log.Print("EDGEDB_CREDENTIALS_FILE environment variable not set")
		return errors.New("credentials not found")
	}

	data, err := ioutil.ReadFile(credFileName)
	if err != nil {
		log.Printf("failed to read credentials file: %q", credFileName)
		return errors.New("credentials not found")
	}

	err = json.Unmarshal(data, &server)
	if err != nil {
		log.Printf("failed to parse credentials file: %q", credFileName)
		return errors.New("credentials not found")
	}
	fmt.Println(server)

	log.Print("using existing server")
	return nil
}

func startServer() (err error) {
	log.Print("starting new server")

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

	server = Options{
		Host:     data.Host,
		Port:     data.Port,
		User:     "edgedb",
		Database: "edgedb",
		admin:    true,
	}

	log.Print("server started")
	return nil
}

func TestMain(m *testing.M) {
	code := 1
	defer func() {
		os.Exit(code)
	}()

	err := getLocalServer()
	if err != nil {
		err = startServer()
		if err != nil {
			panic(err)
		}
	}

	conn, err = Connect(server)
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
