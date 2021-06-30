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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
)

type credentials struct {
	host     string
	port     int
	user     string
	database string
	password string
	certData []byte

	verifyHostname OptionalBool
}

func readCredentials(path string) (*credentials, error) {
	var (
		values map[string]interface{}
		creds  *credentials
	)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		goto Failed
	}

	err = json.Unmarshal(data, &values)
	if err != nil {
		goto Failed
	}

	creds, err = validateCredentials(values)
	if err != nil {
		goto Failed
	}

	return creds, nil

Failed:
	msg := fmt.Sprintf("cannot read credentials at %q: %v", path, err)
	return nil, &configurationError{msg: msg}
}

func validateCredentials(data map[string]interface{}) (*credentials, error) {
	result := &credentials{}

	if val, ok := data["port"]; ok {
		port, ok := val.(float64)
		if !ok || port != math.Trunc(port) || port < 1 || port > 65535 {
			return nil, errors.New("invalid `port` value")
		}
		result.port = int(port)
	} else {
		result.port = 5656
	}

	user, ok := data["user"]
	if !ok {
		return nil, errors.New("`user` key is required")
	}
	result.user, ok = user.(string)
	if !ok {
		return nil, errors.New("`user` must be a string")
	}

	if host, ok := data["host"]; ok {
		result.host, ok = host.(string)
		if !ok {
			return nil, errors.New("`host` must be a string")
		}
	}

	if database, ok := data["database"]; ok {
		result.database, ok = database.(string)
		if !ok {
			return nil, errors.New("`database` must be a string")
		}
	}

	if password, ok := data["password"]; ok {
		result.password, ok = password.(string)
		if !ok {
			return nil, errors.New("`password` must be a string")
		}
	}

	if certData, ok := data["tls_cert_data"]; ok {
		str, ok := certData.(string)
		if !ok {
			return nil, errors.New("`tls_cert_data` must be a string")
		}
		result.certData = []byte(str)
	}

	if verifyHostname, ok := data["tls_verify_hostname"]; ok {
		val, ok := verifyHostname.(bool)
		if !ok {
			return nil, errors.New("`tls_verify_hostname` must be a boolean")
		}

		result.verifyHostname.Set(val)
	}

	return result, nil
}
