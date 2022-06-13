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

	"github.com/edgedb/edgedb-go/internal"
)

// initialized by TestMain
var (
	opts            Options
	client          *Client
	protocolVersion internal.ProtocolVersion
)

func executeOrPanic(command string) {
	ctx := context.Background()
	err := client.Execute(ctx, command)
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
		serverBin = "edgedb-server"
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

	args = append(
		args,
		"--temp-dir",
		"--testmode",
		"--port=auto",
		"--emit-server-status="+statusFileUnix,
		"--tls-cert-mode=generate_self_signed",
		"--auto-shutdown-after=0",
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
		Host:     "127.0.0.1",
		Port:     info.Port,
		User:     "test",
		Password: NewOptionalStr("shhh"),
		Database: "edgedb",
		TLSOptions: TLSOptions{
			CAFile:       info.TLSCertFile,
			SecurityMode: TLSModeNoHostVerification,
		},
	}

	log.Print("server started")
}

func TestMain(m *testing.M) {
	var err error
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
	client, err = CreateClient(ctx, opts)
	if err != nil {
		panic(err)
	}
	log.Println("connected")

	defer client.Close() // nolint:errcheck
	err = client.Execute(ctx,
		"configure instance set session_idle_timeout := <duration>'1s'")
	if err != nil {
		panic(err)
	}
	conn, err := client.acquire(ctx)
	if err != nil {
		panic(err)
	}
	protocolVersion = conn.conn.protocolVersion
	err = client.release(conn, nil)
	if err != nil {
		panic(err)
	}

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
	executeOrPanic("CONFIGURE INSTANCE RESET Auth;")
	executeOrPanic(`
			CONFIGURE INSTANCE INSERT Auth {
				comment := "no password",
				priority := 1,
				method := (INSERT Trust),
				user := {'*'},
			};
		`)
	executeOrPanic(`
			CONFIGURE INSTANCE INSERT Auth {
				comment := "password required",
				priority := 0,
				method := (INSERT SCRAM),
				user := {'user_with_password'}
			}
		`)

	rand.Seed(time.Now().Unix())

	// Some tests explicitly wait for the session idle timeout to expire.
	// When this happens the server will immediately shutdown if there are no
	// active connections. Start a background go routine that keeps an active
	// connection to the database while the tests run so that the server
	// doesn't shutdown.
	done := make(chan struct{}, 1)
	go func() {
		var result int64
		for {
			select {
			case <-done:
				return
			default:
				_ = client.QuerySingle(ctx, "SELECT 1", &result)
			}
		}
	}()
	defer func() { done <- struct{}{} }()

	log.Println("starting tests")
	code = m.Run()
}
