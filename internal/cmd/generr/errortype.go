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

package main

import "strings"

type errorType struct {
	code      [4]uint8
	name      string
	ancestors []string
	tags      []errorTag
}

func (t *errorType) privateName() string {
	return strings.ToLower(t.name[0:1]) + t.name[1:]
}

func parseType(typ []interface{}, lookup map[string]string) *errorType {
	name := typ[0].(string)
	errType := &errorType{
		code: [4]uint8{
			uint8(typ[2].(float64)),
			uint8(typ[3].(float64)),
			uint8(typ[4].(float64)),
			uint8(typ[5].(float64)),
		},
		name: name,
	}

	for _, tag := range typ[6].([]interface{}) {
		errType.tags = append(errType.tags, errorTag(tag.(string)))
	}

	parent := lookup[name]
	for parent != "" {
		errType.ancestors = append(errType.ancestors, parent)
		parent = lookup[parent]
	}

	return errType
}

func parseTypes(data [][]interface{}) []*errorType {
	lookup := make(map[string]string, len(data))
	for _, t := range data {
		name := t[0].(string)
		if !strings.HasSuffix(name, "Error") {
			continue
		}

		parent, _ := t[1].(string)
		lookup[name] = parent
	}

	types := make([]*errorType, 0, len(data))
	for _, d := range data {
		typ := parseType(d, lookup)
		if !strings.HasSuffix(typ.name, "Error") {
			continue
		}

		types = append(types, parseType(d, lookup))
	}

	return types
}
