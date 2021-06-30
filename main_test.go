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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"runtime"
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

type serverInfo struct {
	TLSCertFile string `json:"tls_cert_file"`
	Port        int    `json:"port"`
}

func getServerInfo(fileName string) (*serverInfo, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint:errcheck

	var line string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = scanner.Text()
		if strings.HasPrefix(line, "READY=") {
			break
		}
	}

	err = scanner.Err()
	if err != nil {
		return nil, err
	}

	if line == "" {
		return nil, errors.New("no data found in " + fileName)
	}

	var info serverInfo
	line = strings.TrimPrefix(line, "READY=")
	err = json.Unmarshal([]byte(line), &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// convert a windows path to a unix path for systems with WSL.
func getWSLPath(path string) string {
	path = strings.ReplaceAll(path, "C:", "/mnt/c")
	path = strings.ReplaceAll(path, `\`, "/")
	path = strings.ToLower(path)

	return path
}

func startServer() {
	log.Print("starting new server")

	serverBin := os.Getenv("EDGEDB_SERVER_BIN")
	if serverBin == "" {
		log.Fatal("EDGEDB_SERVER_BIN not set")
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}

	statusFile := path.Join(dir, "status-file")
	log.Println("status file:", dir)

	statusFileUnix := getWSLPath(statusFile)

	args := []string{serverBin}
	if runtime.GOOS == "windows" {
		args = append([]string{"wsl", "-u", "edgedb"}, args...)
	}

	helpArgs := args
	helpArgs = append(helpArgs, "--help")
	out, err := exec.Command(helpArgs[0], helpArgs[1:]...).Output()
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(string(out), "--generate-self-signed-cert") {
		args = append(args, "--generate-self-signed-cert")
	}

	if strings.Contains(string(out), "--auto-shutdown-after") {
		args = append(args, "--auto-shutdown-after=0")
	} else {
		args = append(args, "--auto-shutdown")
	}

	args = append(
		args,
		"--temp-dir",
		"--testmode",
		"--emit-server-status="+statusFileUnix,
		"--port=auto",
		`--bootstrap-command=`+
			`CREATE SUPERUSER ROLE test { SET password := "shhh" }`,
	)

	log.Println("starting server with:", strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)

	cmd.Stderr = os.Stderr
	if os.Getenv("EDGEDB_DEBUG_SERVER") != "" {
		cmd.Stdout = os.Stdout
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	var info *serverInfo
	for i := 0; i < 250; i++ {
		info, err = getServerInfo(statusFile)
		if err == nil && info != nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		cmd.Process.Kill() // nolint:errcheck
		log.Fatal(err)
	}

	if len(info.TLSCertFile) != 0 && runtime.GOOS == "windows" {
		tmpFile := path.Join(dir, "edbtlscert.pem")
		_, err = exec.Command(
			"wsl", "-u", "edgedb", "cp", info.TLSCertFile, getWSLPath(tmpFile),
		).Output()
		if err != nil {
			log.Fatal(err)
		}
		info.TLSCertFile = tmpFile
	}

	opts = Options{
		Hosts:             []string{"127.0.0.1"},
		Ports:             []int{info.Port},
		User:              "test",
		Password:          "shhh",
		Database:          "edgedb",
		TLSCAFile:         info.TLSCertFile,
		TLSVerifyHostname: NewOptionalBool(false),
	}

	log.Print("server started")
}

func TestMain(m *testing.M) {
	var err error = nil
	code := 1
	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			fmt.Println(string(debug.Stack()))
		}

		if err != nil {
			log.Println("error while cleaning up: ", err)
		}
		os.Exit(code)
	}()

	startServer()

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
