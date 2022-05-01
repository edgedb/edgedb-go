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
)

func printCategories(types []*errorType) {
	fmt.Print(`

const (`)

	for _, typ := range types {
		fmt.Printf(`
	%[1]v ErrorCategory = "errors::%[1]v"`, typ.name)
	}

	fmt.Print(`
)`)
}

func printError(errType *errorType) {
	fmt.Printf(`

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
`, errType.name, errType.privateName())

	fmt.Printf(`

func (e *%v) Category(c ErrorCategory) bool {
	switch c {
	case %v:
		return true`, errType.privateName(), errType.name)

	for _, ancestor := range errType.ancestors {
		fmt.Printf(`
	case %v:
		return true`, ancestor)
	}

	fmt.Print(`
	default:
		return false
	}
}
`)
	for _, ancestor := range errType.ancestors {
		fmt.Printf(`
func (e *%v) isEdgeDB%v() {}
`, errType.privateName(), ancestor)
	}

	fmt.Printf(`
func (e *%v) HasTag(tag ErrorTag) bool {
	switch tag {`, errType.privateName())

	for _, tag := range errType.tags {
		fmt.Printf(`
	case %v:
		return true`, tag.identifyer())
	}

	fmt.Printf(`
	default:
		return false
	}
}`)
}

func printErrors(types []*errorType) {
	for _, typ := range types {
		printError(typ)
	}
}

func printCodeMap(types []*errorType) {
	fmt.Print(`

func errorFromCode(code uint32, msg string) error {
	switch code {`)

	for _, typ := range types {
		fmt.Printf(`
	case 0x%02x_%02x_%02x_%02x:
		return &%v{msg: msg}`,
			typ.code[0], typ.code[1], typ.code[2], typ.code[3],
			typ.privateName(),
		)
	}
	code := `
	default:
		return &unexpectedMessageError{
			msg: fmt.Sprintf("invalid error code 0x%x", code),
		}
	}
}`
	fmt.Print(code)
}

func printTags(tags []errorTag) {
	fmt.Print(`

const (`)

	for _, tag := range tags {
		fmt.Printf(`
	%[1]v ErrorTag = %[2]q`, tag.identifyer(), tag)
	}

	fmt.Print(`
)`)
}

func main() {
	var data [][]interface{}
	if e := json.NewDecoder(os.Stdin).Decode(&data); e != nil {
		log.Fatal(e)
	}

	types := parseTypes(data)
	tags := parseTags(data)

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
	fmt.Print(`import "fmt"`)
	printTags(tags)
	printCategories(types)
	printErrors(types)
	printCodeMap(types)
}
