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

//go:build gendocs

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

var lintMode = flag.Bool("lint", false, "Instead of writing output files, "+
	"check if contents of existing files match")

func main() {
	flag.Parse()

	if err := os.Mkdir("rstdocs", 0750); err != nil && !os.IsExist(err) {
		panic(err)
	}

	renderIndexPage()
	typeNames := renderTypesPage()
	renderAPIPage(typeNames)
	renderCodegenPage()
}

func readAndParseFile(fset *token.FileSet, filename string) *ast.File {
	src, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	ast, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	return ast
}

func writeFile(filename string, content string) {
	if *lintMode {
		if file, err := os.ReadFile(filename); err != nil ||
			string(file) != content {
			panic("Content of " + filename + " does not match generated " +
				"docs, Run 'make gendocs' to update docs")
		}
	} else {
		if err := os.WriteFile(
			filename, []byte(content), 0666); err != nil {
			panic(err)
		}
	}
}

func renderTypesPage() []string {
	dir, err := os.ReadDir("internal/geltypes")
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	files := []*ast.File{}

	for _, file := range dir {
		if !file.IsDir() {
			files = append(
				files, readAndParseFile(
					fset, "internal/geltypes/"+file.Name()))
		}
	}

	p, err := doc.NewFromFiles(
		fset, files, "github.com/geldata/gel-go/internal/geltypes")
	if err != nil {
		panic(err)
	}

	rst := `
Datatypes
=========`

	rst += renderTypes(fset, p, p.Types)

	writeFile("rstdocs/types.rst", rst)

	typeNames := []string{}
	for _, t := range p.Types {
		typeNames = append(typeNames, t.Name)
	}

	return typeNames
}

func renderAPIPage(skipTypeNames []string) {
	fset := token.NewFileSet()
	files := []*ast.File{
		readAndParseFile(fset, "export.go"),
	}

	p, err := doc.NewFromFiles(fset, files, "github.com/geldata/gel-go")
	if err != nil {
		panic(err)
	}

	rst := `
API
===`

	skip := make(map[string]bool)
	for _, name := range skipTypeNames {
		skip[name] = true
	}

	types := []*doc.Type{}
	for _, t := range p.Types {
		if !skip[t.Name] {
			types = append(types, t)
		}
	}

	rst += renderTypes(fset, p, types)

	writeFile("rstdocs/api.rst", rst)
}

func renderTypes(
	fset *token.FileSet,
	p *doc.Package,
	types []*doc.Type) string {
	out := ""

	for _, t := range types {
		out += fmt.Sprintf(`


*type* %s
-------%s

`, t.Name, strings.Repeat("-", len(t.Name)))

		out += string(printRST(p.Printer(), p.Parser().Parse(t.Doc)))

		var buf bytes.Buffer
		if err := format.Node(&buf, fset, t.Decl); err != nil {
			panic(err)
		}

		out += "\n.. code-block:: go\n\n    " + strings.ReplaceAll(
			buf.String(), "\n", "\n    ")

		for _, f := range t.Funcs {
			out += fmt.Sprintf(`


*function* %s
...........%s
`, f.Name, strings.Repeat(".", len(f.Name)))

			var buf bytes.Buffer
			if err := format.Node(&buf, fset, f.Decl); err != nil {
				panic(err)
			}

			out += "\n.. code-block:: go\n\n    " + strings.ReplaceAll(
				buf.String(), "\n", "\n    ") + "\n\n"

			out += string(printRST(p.Printer(), p.Parser().Parse(f.Doc)))
		}

		for _, m := range t.Methods {
			out += fmt.Sprintf(`


*method* %s
.........%s
`, m.Name, strings.Repeat(".", len(m.Name)))

			var buf bytes.Buffer
			if err := format.Node(&buf, fset, m.Decl); err != nil {
				panic(err)
			}

			out += "\n.. code-block:: go\n\n    " + strings.ReplaceAll(
				buf.String(), "\n", "\n    ") + "\n\n"

			out += string(printRST(p.Printer(), p.Parser().Parse(m.Doc)))
		}
	}

	return strings.ReplaceAll(out, "\t", "    ")
}

func renderIndexPage() {
	fset := token.NewFileSet()
	files := []*ast.File{
		readAndParseFile(fset, "doc.go"),
		readAndParseFile(fset, "doc_test.go"),
	}

	p, err := doc.NewFromFiles(fset, files, "github.com/geldata/gel-go")
	if err != nil {
		panic(err)
	}

	rst := `.. _edgedb-go-intro:

=============
Gel Go Driver
=============


.. toctree::
   :maxdepth: 3
   :hidden:

   api
   types
   codegen



`

	rst += string(printRST(p.Printer(), p.Parser().Parse(p.Doc)))

	rst += `

Usage Example
-------------

.. code-block:: go
`

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, p.Examples[0].Code); err != nil {
		panic(err)
	}

	exampleLines := strings.Split(buf.String(), "\n")

	skip := true
	for _, line := range exampleLines {
		if skip && !strings.HasPrefix(line, "//") {
			skip = false
		}
		if !skip {
			rst += "    " + strings.ReplaceAll(line, "\t", "    ") + "\n"
		}
	}

	writeFile("rstdocs/index.rst", rst)
}

func renderCodegenPage() {
	fset := token.NewFileSet()
	files := []*ast.File{
		readAndParseFile(fset, "cmd/edgeql-go/doc.go"),
	}

	p, err := doc.NewFromFiles(
		fset, files, "github.com/geldata/gel-go/cmd/edgeql-go")
	if err != nil {
		panic(err)
	}

	rst := `
Codegen
=======
`

	rst += string(printRST(p.Printer(), p.Parser().Parse(p.Doc)))

	writeFile("rstdocs/codegen.rst", rst)
}
