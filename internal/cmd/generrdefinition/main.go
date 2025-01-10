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

//go:build tools

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/geldata/gel-go/internal/errgen"
)

func printCategories(types []*errgen.Type) {
	fmt.Print(`

const (`)

	for _, typ := range types {
		fmt.Printf(`
	%[1]v ErrorCategory = "errors::%[1]v"`, typ.Name)
	}

	fmt.Print(`
)`)
}

func printError(errType *errgen.Type) {
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

	return "gel.%[1]v: " + msg
}

func (e *%[2]v) Unwrap() error { return e.err }
`, errType.Name, errType.PrivateName())

	fmt.Printf(`

func (e *%v) Category(c ErrorCategory) bool {
	switch c {
	case %v:
		return true`, errType.PrivateName(), errType.Name)

	for _, ancestor := range errType.Ancestors {
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
	for _, ancestor := range errType.Ancestors {
		fmt.Printf(`
func (e *%v) isEdgeDB%v() {}
`, errType.PrivateName(), ancestor)
	}

	fmt.Printf(`
func (e *%v) HasTag(tag ErrorTag) bool {
	switch tag {`, errType.PrivateName())

	for _, tag := range errType.Tags {
		fmt.Printf(`
	case %v:
		return true`, tag.Identifyer())
	}

	fmt.Printf(`
	default:
		return false
	}
}`)
}

func printErrors(types []*errgen.Type) {
	for _, typ := range types {
		printError(typ)
	}
}

func printCodeMap(types []*errgen.Type) {
	fmt.Print(`

func errorFromCode(code uint32, msg string) error {
	switch code {`)

	for _, typ := range types {
		fmt.Printf(`
	case 0x%02x_%02x_%02x_%02x:
		return &%v{msg: msg}`,
			typ.Code[0], typ.Code[1], typ.Code[2], typ.Code[3],
			typ.PrivateName(),
		)
	}
	code := `
	default:
		return &unexpectedMessageError{
			msg: fmt.Sprintf(
				"invalid error code 0x%x with message %q", code, msg,
			),
		}
	}
}`
	fmt.Print(code)
}

func printTags(tags []errgen.Tag) {
	fmt.Print(`

const (`)

	for _, tag := range tags {
		fmt.Printf(`
	%[1]v ErrorTag = %[2]q`, tag.Identifyer(), tag)
	}

	fmt.Print(`
)`)
}

//nolint:typecheck
func main() {
	var data [][]interface{}
	if e := json.NewDecoder(os.Stdin).Decode(&data); e != nil {
		log.Fatal(e)
	}

	types := errgen.ParseTypes(data)
	tags := errgen.ParseTags(data)

	fmt.Print(`// This source file is part of the EdgeDB open source project.
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

// This file is auto generated. Do not edit!
// run 'make errors' to regenerate

// internal/cmd/export should ignore this file
//go:build !export
`)

	fmt.Println()
	fmt.Println("package gel")
	fmt.Println()
	fmt.Print(`import "fmt"`)
	printTags(tags)
	printCategories(types)
	printErrors(types)
	printCodeMap(types)
}
