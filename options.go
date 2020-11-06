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
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Options for connecting to an EdgeDB server
type Options struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Database string `json:"database"`
	Password string `json:"password"`
	admin    bool
}

func (o *Options) network() string {
	if o.admin {
		return "unix"
	}
	return "tcp"
}

func (o *Options) address() string {
	if o.admin {
		return fmt.Sprintf("%v/.s.EDGEDB.admin.%v", o.Host, o.Port)
	}

	host := o.Host
	if host == "" {
		host = "localhost"
	}

	port := o.Port
	if port == 0 {
		port = 5656
	}

	return fmt.Sprintf("%v:%v", host, port)
}

// DSN parses a URI string into an Options struct
func DSN(dsn string) (opts Options, err error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return opts, err
	}

	if parsed.Scheme != "edgedb" {
		return opts, fmt.Errorf("dsn %q is not an edgedb:// URI", dsn)
	}

	var port int
	if parsed.Port() == "" {
		port = 5656
	} else {
		port, err = strconv.Atoi(parsed.Port())
		if err != nil {
			return opts, err
		}
	}

	host := strings.Split(parsed.Host, ":")[0]
	db := strings.TrimLeft(parsed.Path, "/")
	password, _ := parsed.User.Password()

	return Options{
		Host:     host,
		Port:     port,
		User:     parsed.User.Username(),
		Database: db,
		Password: password,
	}, nil
}
