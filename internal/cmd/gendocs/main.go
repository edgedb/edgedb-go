package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func main() {

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

func renderTypesPage() []string {
	dir, err := os.ReadDir("internal/edgedbtypes")
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	files := []*ast.File{}

	for _, file := range dir {
		if !file.IsDir() {
			files = append(
				files, readAndParseFile(fset, "internal/edgedbtypes/"+file.Name()))
		}
	}

	p, err := doc.NewFromFiles(
		fset, files, "github.com/edgedb/edgedb-go/internal/edgedbtypes")
	if err != nil {
		panic(err)
	}

	rst := `
Datatypes
=========`

	rst += renderTypes(fset, p, p.Types)

	if err := os.WriteFile("rstdocs/types.rst", []byte(rst), 0666); err != nil {
		panic(err)
	}

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

	p, err := doc.NewFromFiles(fset, files, "github.com/edgedb/edgedb-go")
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

	if err := os.WriteFile("rstdocs/api.rst", []byte(rst), 0666); err != nil {
		panic(err)
	}
}

func renderTypes(fset *token.FileSet, p *doc.Package, types []*doc.Type) string {
	out := ""

	for _, t := range types {
		out += fmt.Sprintf(`


*type* %s
-------%s

`, t.Name, strings.Repeat("-", len(t.Name)))

		out += string(printRST(p.Printer(), p.Parser().Parse(t.Doc)))

		var buf bytes.Buffer
		format.Node(&buf, fset, t.Decl)

		out += "\n.. code-block:: go\n\n    " + strings.ReplaceAll(
			buf.String(), "\n", "\n    ")

		for _, f := range t.Funcs {
			out += fmt.Sprintf(`


*function* %s
...........%s
`, f.Name, strings.Repeat(".", len(f.Name)))

			var buf bytes.Buffer
			format.Node(&buf, fset, f.Decl)

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
			format.Node(&buf, fset, m.Decl)

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

	p, err := doc.NewFromFiles(fset, files, "github.com/edgedb/edgedb-go")
	if err != nil {
		panic(err)
	}

	rst := `
================
EdgeDB Go Driver
================


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
	format.Node(&buf, fset, p.Examples[0].Code)

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

	if err := os.WriteFile("rstdocs/index.rst", []byte(rst), 0666); err != nil {
		panic(err)
	}
}

func renderCodegenPage() {
	fset := token.NewFileSet()
	files := []*ast.File{
		readAndParseFile(fset, "cmd/edgeql-go/doc.go"),
	}

	p, err := doc.NewFromFiles(
		fset, files, "github.com/edgedb/edgedb-go/cmd/edgeql-go")
	if err != nil {
		panic(err)
	}

	rst := `
Codegen
=======
`

	rst += string(printRST(p.Printer(), p.Parser().Parse(p.Doc)))

	if err := os.WriteFile("rstdocs/codegen.rst", []byte(rst), 0666); err != nil {
		panic(err)
	}
}
