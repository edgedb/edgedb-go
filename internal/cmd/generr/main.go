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

type typeList [][]interface{}

func printTypes(types typeList) {
	for i, t := range types {
		name := t[0].(string)

		fmt.Printf("// %v is an error.\n", name)
		fmt.Printf("type %[1]v struct {\n", name)
		fmt.Printf("\t*baseError\n")
		fmt.Printf("}\n")

		if i < len(types)-1 {
			fmt.Println()
		}
	}
}

func codeFromName(name string) string {
	name = strings.ToLower(name[0:1]) + name[1:]
	return name + "Code"
}

func printCodes(types typeList) {
	fmt.Println("const (")

	for _, t := range types {
		name := t[0].(string)
		code := codeFromName(name)

		fmt.Printf("\t%v uint32 = 0x%02x_%02x_%02x_%02x\n",
			code,
			int(t[2].(float64)),
			int(t[3].(float64)),
			int(t[4].(float64)),
			int(t[5].(float64)),
		)
	}

	fmt.Println(")")
}

func printTree(types typeList) {
	fmt.Println("// wrapErrorWithType wraps an error in an edgedb error type.")
	fmt.Println("func wrapErrorWithType(code uint32, err error) error {")
	fmt.Println("\tif err == nil {")
	fmt.Println("\t\treturn nil")
	fmt.Println("\t}")
	fmt.Println()
	fmt.Println("\tswitch code {")

	for _, t := range types {
		name := t[0].(string)
		code := codeFromName(name)

		switch parent := t[1].(type) {
		case string:
			pCode := codeFromName(parent)
			fmt.Printf("\tcase %v:\n", code)
			fmt.Printf("\t\tnext := wrapErrorWithType(%v, err)\n", pCode)
			fmt.Printf("\t\treturn &%v{&baseError{err: next}}\n", name)
		case nil:
			fmt.Printf("\tcase %v:\n", code)
			fmt.Printf(
				"\t\treturn &%v{&baseError{err: wrapError(err)}}\n", name)
		default:
			panic("unexpected type")
		}
	}

	fmt.Print("\tdefault:\n")
	fmt.Printf("\t\tpanic(fmt.Sprintf(\"unknown error code: %%v\", code))\n")
	fmt.Print("\t}\n")
	fmt.Print("}\n")
}

// log messages and warnings don't make sense as error types
func filterNonError(types typeList) typeList {
	newTypes := typeList{}

	for _, t := range types {
		name := t[0].(string)
		if !strings.HasSuffix(name, "Error") {
			continue
		}
		newTypes = append(newTypes, t)
	}

	return newTypes
}

func main() {
	var types typeList
	if e := json.NewDecoder(os.Stdin).Decode(&types); e != nil {
		log.Fatal(e)
	}

	types = filterNonError(types)

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
`)

	fmt.Println()
	fmt.Println("package edgedb")
	fmt.Println()
	fmt.Println(`import "fmt"`)
	fmt.Println()
	printCodes(types)
	fmt.Println()
	printTree(types)
	fmt.Println()
	printTypes(types)
}
