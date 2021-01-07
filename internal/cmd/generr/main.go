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

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

func private(name string) string {
	return strings.ToLower(name[0:1]) + name[1:]
}

func printError(name, parent string, ancestors []string) {
	fmt.Printf(`
// %[1]v is an error.
type %[1]v interface {
	%[3]v
	is%[1]v()
}

type %[2]v struct {
	msg string
	err error
}

func (e *%[2]v) Error() string {
	msg := e.msg
	if e.err != nil {
		msg = e.err.Error()
	}

	return "edgedb.%[1]v: " + msg
}

func (e *%[2]v) Unwrap() error { return e.err }

func (e *%[2]v) is%[1]v() {}

func (e *%[2]v) is%[3]v() {}
`, name, private(name), parent)

	for _, ancestor := range ancestors {
		fmt.Printf(`
func (e *%v) is%v() {}
`, private(name), ancestor)
	}
}

func printErrors(types [][]string) {
	for _, lineage := range types {
		name := lineage[0]
		parent := lineage[1]
		printError(name, parent, lineage[2:])
	}
}

func printCodeMap(data [][]interface{}) {
	fmt.Print(`
func errorFromCode(code uint32, msg string) error {
	switch code {`)

	for _, t := range data {
		fmt.Printf(`
	case 0x%02x_%02x_%02x_%02x:
		return &%v{msg: msg}`,
			int(t[2].(float64)),
			int(t[3].(float64)),
			int(t[4].(float64)),
			int(t[5].(float64)),
			private(t[0].(string)),
		)
	}

	fmt.Print(`
	default:
		panic(fmt.Sprintf("invalid error code 0x%` + `x", code))
	}
}
`)
}

func main() {
	var data [][]interface{}
	if e := json.NewDecoder(os.Stdin).Decode(&data); e != nil {
		log.Fatal(e)
	}

	lookup := make(map[string]string, len(data))
	for _, t := range data {
		name := t[0].(string)
		if !strings.HasSuffix(name, "Error") {
			continue
		}

		parent, _ := t[1].(string)
		lookup[name] = parent
	}

	types := make([][]string, 0, len(lookup))
	for _, t := range data {
		name := t[0].(string)
		parent := lookup[name]
		parents := []string{name}

		for parent != "" {
			parents = append(parents, parent)
			parent = lookup[parent]
		}

		parents = append(parents, "Error")
		types = append(types, parents)
	}

	fmt.Print(`// This source file is part of the EdgeDB open source project.
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

// This file is auto generated. Do not edit!
// run 'make errors' to regenerate
`)

	fmt.Println()
	fmt.Println("package edgedb")
	fmt.Println()
	fmt.Println(`import "fmt"`)
	printErrors(types)
	printCodeMap(data)
}
