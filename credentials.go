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
}

func readCredentials(path string) (*credentials, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot read credentials at %q: %v%w", path, err, ErrBadConfig,
		)
	}

	var values map[string]interface{}
	if e := json.Unmarshal(data, &values); e != nil {
		return nil, fmt.Errorf(
			"cannot read credentials at %q: %v%w", path, e, ErrBadConfig,
		)
	}

	creds, err := validateCredentials(values)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot read credentials at %q: %v%w", path, err, ErrBadConfig,
		)
	}

	return creds, nil
}

func validateCredentials(data map[string]interface{}) (*credentials, error) {
	result := &credentials{}

	if val, ok := data["port"]; ok {
		port, ok := val.(float64)
		if !ok || port != math.Trunc(port) || port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid `port` value%w", ErrBadConfig)
		}
		result.port = int(port)
	} else {
		result.port = 5656
	}

	user, ok := data["user"]
	if !ok {
		return nil, fmt.Errorf("`user` key is required%w", ErrBadConfig)
	}
	result.user, ok = user.(string)
	if !ok {
		return nil, fmt.Errorf("`user` must be a string%w", ErrBadConfig)
	}

	if host, ok := data["host"]; ok {
		result.host, ok = host.(string)
		if !ok {
			return nil, fmt.Errorf("`host` must be a string%w", ErrBadConfig)
		}
	}

	if database, ok := data["database"]; ok {
		result.database, ok = database.(string)
		if !ok {
			return nil, fmt.Errorf(
				"`database` must be a string%w", ErrBadConfig,
			)
		}
	}

	if password, ok := data["password"]; ok {
		result.password, ok = password.(string)
		if !ok {
			return nil, fmt.Errorf(
				"`password` must be a string%w", ErrBadConfig,
			)
		}
	}

	return result, nil
}
