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
	"os"
	usr "os/user"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const edgedbPort = 5656

type connConfig struct {
	addrs          []dialArgs
	user           string
	password       string
	database       string
	connectTimeout time.Duration
	serverSettings map[string]string
}

type dialArgs struct {
	network string
	address string
}

func validatePortSpec(hosts []string, ports []int) ([]int, error) {
	var result []int
	if len(ports) > 1 {
		if len(ports) != len(hosts) {
			return nil, newErrorFromCode(configurationErrorCode, fmt.Sprintf(
				"could not match %v port numbers to %v hosts",
				len(ports), len(hosts),
			))
		}

		result = ports
	} else {
		result = make([]int, len(hosts))
		for i := 0; i < len(hosts); i++ {
			result[i] = ports[0]
		}
	}

	return result, nil
}

func parsePortSpec(spec string) ([]int, error) {
	ports := make([]int, 0, strings.Count(spec, ","))

	for _, p := range strings.Split(spec, ",") {
		port, err := strconv.Atoi(p)
		if err != nil {
			return nil, newErrorFromCode(configurationErrorCode, fmt.Sprintf(
				"invalid port %q found in %q: %v",
				p, spec, err,
			))
		}

		ports = append(ports, port)
	}

	return ports, nil
}

func parseHostList(hostList string, ports []int) ([]string, []int, error) {
	hostSpecs := strings.Split(hostList, ",")

	var (
		err           error
		defaultPorts  []int
		hostListPorts []int
	)

	if len(ports) == 0 {
		if portSpec := os.Getenv("EDGEDB_PORT"); portSpec != "" {
			defaultPorts, err = parsePortSpec(portSpec)
			if err != nil {
				return nil, nil, err
			}
		} else {
			defaultPorts = []int{edgedbPort}
		}

		defaultPorts, err = validatePortSpec(hostSpecs, defaultPorts)
		if err != nil {
			return nil, nil, err
		}
	} else {
		ports, err = validatePortSpec(hostSpecs, ports)
		if err != nil {
			return nil, nil, err
		}
	}

	hosts := make([]string, 0, len(hostSpecs))
	for i, hostSpec := range hostSpecs {
		addr, hostSpecPort := partition(hostSpec, ":")
		hosts = append(hosts, addr)

		if len(ports) == 0 {
			if hostSpecPort != "" {
				port, err := strconv.Atoi(hostSpecPort)
				if err != nil {
					msg := fmt.Sprintf(
						"invalid port %q found in %q: %v",
						hostSpecPort, hostSpec, err,
					)
					err = newErrorFromCode(configurationErrorCode, msg)
					return nil, nil, err
				}
				hostListPorts = append(hostListPorts, port)
			} else {
				hostListPorts = append(hostListPorts, defaultPorts[i])
			}
		}
	}

	if len(ports) == 0 {
		ports = hostListPorts
	}

	return hosts, ports, nil
}

func partition(s, sep string) (string, string) {
	list := strings.SplitN(s, sep, 2)
	switch len(list) {
	case 2:
		return list[0], list[1]
	case 1:
		return list[0], ""
	default:
		return "", ""
	}
}

func pop(m map[string]string, key string) string {
	v, ok := m[key]
	if ok {
		delete(m, key)
	}
	return v
}

func parseConnectDSNAndArgs(
	dsn string,
	opts *Options,
) (*connConfig, error) {
	usingCredentials := false
	hosts := opts.Hosts
	ports := opts.Ports
	user := opts.User
	password := opts.Password
	database := opts.Database

	serverSettings := make(map[string]string, len(opts.ServerSettings))
	for k, v := range opts.ServerSettings {
		serverSettings[k] = v
	}

	if dsn != "" && strings.HasPrefix(dsn, "edgedb://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			return nil, newErrorFromCode(configurationErrorCode, fmt.Sprintf(
				"could not parse %q: %v", dsn, err))
		}

		if parsed.Scheme != "edgedb" {
			return nil, newErrorFromCode(configurationErrorCode, fmt.Sprintf(
				`invalid DSN: scheme is expected to be "edgedb", got %q`, dsn,
			))
		}

		if len(hosts) == 0 && parsed.Host != "" {
			hosts, ports, err = parseHostList(parsed.Host, ports)
			if err != nil {
				return nil, err
			}
		}

		if database == "" {
			database = strings.TrimLeft(parsed.Path, "/")
		}

		if user == "" {
			user = parsed.User.Username()
		}

		if password == "" {
			password, _ = parsed.User.Password()
		}

		if parsed.RawQuery != "" {
			q, err := url.ParseQuery(parsed.RawQuery)
			if err != nil {
				msg := fmt.Sprintf("invalid DSN %q: %v", dsn, err)
				return nil, newErrorFromCode(configurationErrorCode, msg)
			}

			query := make(map[string]string, len(q))
			for key, val := range q {
				query[key] = val[len(val)-1]
			}

			if val := pop(query, "port"); val != "" && len(ports) == 0 {
				ports, err = parsePortSpec(val)
				if err != nil {
					return nil, err
				}
			}

			if val := pop(query, "host"); val != "" && len(hosts) == 0 {
				hosts, ports, err = parseHostList(val, ports)
				if err != nil {
					return nil, err
				}
			}

			if val := pop(query, "dbname"); database == "" {
				database = val
			}

			if val := pop(query, "database"); database == "" {
				database = val
			}

			if val := pop(query, "user"); user == "" {
				user = val
			}

			if val := pop(query, "password"); password == "" {
				password = val
			}

			for k, v := range query {
				serverSettings[k] = v
			}
		}
	} else if dsn != "" {
		isIdentifier := regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)
		if !isIdentifier.Match([]byte(dsn)) {
			return nil, newErrorFromCode(configurationErrorCode, fmt.Sprintf(
				"dsn %q is neither a edgedb:// URI nor valid instance name",
				dsn,
			))
		}

		usingCredentials = true

		u, err := usr.Current()
		if err != nil {
			return nil, newErrorFromCode(configurationErrorCode, err.Error())
		}

		file := path.Join(u.HomeDir, ".edgedb", "credentials", dsn+".json")
		creds, err := readCredentials(file)
		if err != nil {
			return nil, newErrorFromCode(configurationErrorCode, fmt.Sprintf(
				"cannot read credentials of instance %q: %v", dsn, err,
			))
		}

		if len(ports) == 0 {
			ports = []int{creds.port}
		}

		if user == "" {
			user = creds.user
		}

		if len(hosts) == 0 && creds.host != "" {
			hosts = []string{creds.host}
		}

		if password == "" {
			password = creds.password
		}

		if database == "" {
			database = creds.database
		}
	}

	var err error

	if spec := os.Getenv("EDGEDB_HOST"); len(hosts) == 0 && spec != "" {
		hosts, ports, err = parseHostList(spec, ports)
		if err != nil {
			return nil, err
		}
	}

	if len(hosts) == 0 {
		if !usingCredentials {
			hosts = append(hosts, defaultHosts...)
		}
		hosts = append(hosts, "localhost")
	}

	if len(ports) == 0 {
		if portSpec := os.Getenv("EDGEDB_PORT"); portSpec != "" {
			ports, err = parsePortSpec(portSpec)
			if err != nil {
				return nil, err
			}
		} else {
			ports = []int{edgedbPort}
		}
	}

	ports, err = validatePortSpec(hosts, ports)
	if err != nil {
		return nil, err
	}

	if user == "" {
		user = os.Getenv("EDGEDB_USER")
	}

	if user == "" {
		user = "edgedb"
	}

	if password == "" {
		password = os.Getenv("EDGEDB_PASSWORD")
	}

	if database == "" {
		database = os.Getenv("EDGEDB_DATABASE")
	}

	if database == "" {
		database = "edgedb"
	}

	var addrs []dialArgs
	for i := 0; i < len(hosts); i++ {
		h := hosts[i]
		p := ports[i]

		if strings.HasPrefix(h, "/") {
			if !strings.Contains(h, ".s.EDGEDB.") {
				h = path.Join(h, fmt.Sprintf(".s.EDGEDB.%v", p))
			}
			addrs = append(addrs, dialArgs{"unix", h})
		} else {
			addrs = append(addrs, dialArgs{
				"tcp",
				fmt.Sprintf("%v:%v", h, p),
			})
		}
	}

	if len(addrs) == 0 {
		return nil, newErrorFromCode(configurationErrorCode,
			"could not determine the database address to connect to",
		)
	}

	cfg := &connConfig{
		addrs:          addrs,
		user:           user,
		password:       password,
		database:       database,
		connectTimeout: opts.ConnectTimeout,
		serverSettings: serverSettings,
	}

	return cfg, nil
}
