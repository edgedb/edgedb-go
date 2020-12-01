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
	"time"
)

// Options for connecting to an EdgeDB server
type Options struct {
	// Hosts is a slice of database host addresses as one of the following
	//
	// - an IP address or domain name
	//
	// - an absolute path to the directory
	//   containing the database server Unix-domain socket
	//   (not supported on Windows)
	//
	// If the slice is empty, the following will be tried, in order:
	//
	// - host address(es) parsed from the dsn argument
	//
	// - the value of the EDGEDB_HOST environment variable
	//
	// - on Unix, common directories used for EdgeDB Unix-domain sockets:
	//   "/run/edgedb" and "/var/run/edgedb"
	//
	// - "localhost"
	Hosts []string

	// Ports is a slice of port numbers to connect to at the server host
	// (or Unix-domain socket file extension).
	//
	// Ports may either be:
	//
	// - the same length ans Hosts
	//
	// - a single port to be used all specified hosts
	//
	// - empty indicating the value parsed from the dsn argument
	//   should be used, or the value of the EDGEDB_PORT environment variable,
	//   or 5656 if neither is specified.
	Ports []int

	// User is the name of the database role used for authentication.
	// If not specified, the value parsed from the dsn argument is used,
	// or the value of the EDGEDB_USER environment variable,
	// or the operating system name of the user running the application.
	User string

	// Database is the name of the database to connect to.
	// If not specified, the value parsed from the dsn argument is used,
	// or the value of the EDGEDB_DATABASE environment variable,
	// or the operating system name of the user running the application.
	Database string

	// Password to be used for authentication,
	// if the server requires one. If not specified,
	// the value parsed from the dsn argument is used,
	// or the value of the EDGEDB_PASSWORD environment variable.
	// Note that the use of the environment variable is discouraged
	// as other users and applications may be able to read it
	// without needing specific privileges.
	Password string

	// ConnectTimeout is used when establishing connections in the background.
	ConnectTimeout time.Duration

	// MinConns determines the minimum number of connections.
	MinConns int

	// MaxConns determines the maximum number of connections.
	MaxConns int

	ServerSettings map[string]string
}
