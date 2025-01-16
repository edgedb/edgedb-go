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
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/geldata/gel-go/internal"
	"github.com/geldata/gel-go/internal/geltypes"
)

var (
	client         *Client
	once           sync.Once
	testServerInfo = filepath.Join(
		os.TempDir(),
		"edgedb-go-test-server-info",
	)
	testServerConfigured bool

	opts            Options
	protocolVersion internal.ProtocolVersion
)

// TestClient returns a client connected to the test server.
func TestClient() *Client {
	once.Do(initServer)
	return client
}

// TestClientOptions returns the Options used to connect to the test server.
func TestClientOptions() Options {
	once.Do(initServer)
	return opts
}

// TestClientProtocolVersion returns the protocol version used to connect to
// the test server.
func TestClientProtocolVersion() internal.ProtocolVersion {
	once.Do(initServer)
	return protocolVersion
}

func fatal(err error) {
	debug.PrintStack()
	log.Fatal(err)
}

func execOrFatal(command string) {
	ctx := context.Background()
	err := client.Execute(ctx, command)
	if err != nil {
		fatal(err)
	}
}

func initServer() {
	defer log.Println("test server is ready for use")

	initServerInfo()
	initClient()
	initProtocolVersion()

	if testServerConfigured {
		return
	}

	log.Println("configuring instance")
	execOrFatal(`
		CONFIGURE INSTANCE SET session_idle_timeout := <duration>'1s';
	`)
	execOrFatal(`
		CREATE SUPERUSER ROLE user_with_password {
			SET password := 'secret';
		};
	`)
	execOrFatal(`
		CONFIGURE INSTANCE RESET Auth;
	`)
	execOrFatal(`
		CONFIGURE INSTANCE INSERT Auth {
			comment := "no password",
			priority := 1,
			method := (INSERT Trust),
			user := {'*'},
		};
	`)
	execOrFatal(`
		CONFIGURE INSTANCE INSERT Auth {
			comment := "password required",
			priority := 0,
			method := (INSERT SCRAM),
			user := {'user_with_password'}
		};
	`)

	log.Println("running migration")
	if protocolVersion.GTE(protocolVersion1p0) {
		execOrFatal(`
			START MIGRATION TO {
				module default {
					global global_id -> uuid;
					required global global_str -> str {
						default := "default";
					};
					global global_bytes -> bytes;
					global global_int16 -> int16;
					global global_int32 -> int32;
					global global_int64 -> int64;
					global global_float32 -> float32;
					global global_float64 -> float64;
					global global_bool -> bool;
					global global_datetime -> datetime;
					global global_duration -> duration;
					global global_json -> json;
					global global_local_datetime -> cal::local_datetime;
					global global_local_date -> cal::local_date;
					global global_local_time -> cal::local_time;
					global global_bigint -> bigint;
					global global_relative_duration -> cal::relative_duration;
					global global_date_duration -> cal::date_duration;
					global global_memory -> cfg::memory;

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
	} else {
		execOrFatal(`
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
	}
}

func initServerInfo() {
	if err := readServerInfo(); err == nil {
		return
	}

	startServerProcess()
}

func initOptions(info serverInfo) {
	opts = Options{
		Host:     "127.0.0.1",
		Port:     info.Port,
		User:     "test",
		Password: geltypes.NewOptionalStr("shhh"),
		TLSOptions: TLSOptions{
			CAFile:       info.TLSCertFile,
			SecurityMode: TLSModeNoHostVerification,
		},
	}
}

func readServerInfo() (err error) {
	defer func() {
		if err != nil {
			opts = Options{}
			log.Println("error reading test server info file:", err)
		} else {
			testServerConfigured = true
		}
	}()

	var data []byte
	data, err = os.ReadFile(testServerInfo)
	if err != nil {
		return err
	}

	var info serverInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return err
	}

	ctx := context.Background()
	initOptions(info)
	o := opts
	o.WaitUntilAvailable = 500 * time.Millisecond
	c, err := CreateClient(ctx, o)
	if err != nil {
		return err
	}

	err = c.Execute(ctx, "select 1")
	if err != nil {
		return err
	}

	return nil
}

func startServerProcess() {
	log.Print("starting test server")
	defer log.Print("test server started")

	serverBin := os.Getenv("EDGEDB_SERVER_BIN")
	if serverBin == "" {
		serverBin = "edgedb-server"
	}

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		fatal(err)
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
		"--auto-shutdown-after=10",
		`--bootstrap-command=`+
			`CREATE SUPERUSER ROLE test { SET password := "shhh" }`,
	)

	log.Println("starting server with:", strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)

	if os.Getenv("EDGEDB_SILENT_SERVER") == "" {
		fmt.Print(`
-------------------------------------------------------------------------------
Forwarding server's stderr. Set EDGEDB_SILENT_SERVER=1 to suppress.
-------------------------------------------------------------------------------

`)
		cmd.Stderr = os.Stderr
	} else {
		fmt.Print(`
-------------------------------------------------------------------------------
EDGEDB_SILENT_SERVER is set. Hiding server's stderr.
-------------------------------------------------------------------------------

`)
	}

	if os.Getenv("EDGEDB_DEBUG_SERVER") != "" {
		fmt.Print(`
-------------------------------------------------------------------------------
EDGEDB_DEBUG_SERVER is set. Forwarding server's stdout.
-------------------------------------------------------------------------------

`)
		cmd.Stdout = os.Stdout
	} else {
		fmt.Print(`
-------------------------------------------------------------------------------
Set EDGEDB_DEBUG_SERVER=1 to see server debug logs.
-------------------------------------------------------------------------------

`)
	}

	if os.Getenv("CI") == "" && os.Getenv("EDGEDB_SERVER_BIN") == "" {
		cmd.Env = append(os.Environ(),
			"__EDGEDB_DEVMODE=1",
		)
	}

	err = cmd.Start()
	if err != nil {
		fatal(err)
	}

	log.Println("waiting for test server connection info")
	var info *serverInfo
	for i := 0; i < 250; i++ {
		info, err = getServerInfo(statusFile)
		if err == nil && info != nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		_ = cmd.Process.Kill()
		fatal(err)
	}

	if len(info.TLSCertFile) != 0 && runtime.GOOS == "windows" {
		tmpFile := path.Join(dir, "edbtlscert.pem")
		_, err = exec.Command(
			"wsl", "-u", "edgedb", "cp", info.TLSCertFile, getWSLPath(tmpFile),
		).Output()
		if err != nil {
			fatal(err)
		}
		info.TLSCertFile = tmpFile
	}

	data, err := json.Marshal(info)
	if err != nil {
		fatal(err)
	}

	err = os.WriteFile(testServerInfo, data, 0777)
	if err != nil {
		fatal(err)
	}

	initOptions(*info)
}

// convert a windows path to a unix path for systems with WSL.
func getWSLPath(path string) string {
	path = strings.ReplaceAll(path, "C:", "/mnt/c")
	path = strings.ReplaceAll(path, `\`, "/")
	path = strings.ToLower(path)

	return path
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

func initClient() {
	log.Println("initializing testserver.Client")

	ctx := context.Background()
	var err error
	client, err = CreateClient(ctx, opts)
	if err != nil {
		fatal(err)
	}
}

func initProtocolVersion() {
	log.Println("initializing testserver.ProtocolVersion")
	var err error
	protocolVersion, err = ProtocolVersion(context.Background(), client)
	if err != nil {
		fatal(err)
	}
	log.Printf("using Protocol Version: %v", protocolVersion)
}
