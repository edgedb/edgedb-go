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

// export generates exported identifiers and copies their docstring.
package main

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"

	"golang.org/x/exp/constraints"
)

// Export holds the information needed to export an identifier.
type Export struct {
	Tok     token.Token
	PkgName string
	Name    string
	Comment *ast.CommentGroup
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("internal/cmd/export: ")

	edb, err := buildLookup("internal/client", "gel")
	if err != nil {
		log.Fatal(err)
	}

	typ, err := buildLookup("internal/geltypes", "geltypes")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("internal/cmd/export/names.txt")
	if err != nil {
		log.Fatal(err)
	}

	exports := make(map[token.Token][]Export)
	appendExport := func(e Export) {
		if e.Comment == nil {
			log.Fatalf(
				"cannot export %s.%s because it is undocumented\n",
				e.PkgName,
				e.Name,
			)
		}
		exports[e.Tok] = append(exports[e.Tok], e)
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		e, ok := typ[scanner.Text()]
		if ok {
			appendExport(e)
			continue
		}
		e, ok = edb[scanner.Text()]
		if ok {
			appendExport(e)
			continue
		}

		log.Fatalf("%q is not available to export", scanner.Text())
	}

	var buf bytes.Buffer
	t, err := template.ParseGlob("internal/cmd/export/templates/*.template")
	if err != nil {
		log.Fatal(err)
	}

	quicksort(exports[token.CONST], func(e Export) string { return e.Name })
	quicksort(exports[token.TYPE], func(e Export) string { return e.Name })
	quicksort(exports[token.VAR], func(e Export) string { return e.Name })
	err = t.Execute(&buf, map[string]any{
		"PackageName": "gel",
		"Imports": []string{
			`gel "github.com/geldata/gel-go/internal/client"`,
			`"github.com/geldata/gel-go/internal/geltypes"`,
		},
		"Constants": exports[token.CONST],
		"Types":     exports[token.TYPE],
		"Vars":      exports[token.VAR],
	})
	if err != nil {
		log.Fatal(err)
	}

	o, err := os.Create("export.go")
	if err != nil {
		log.Fatal(err)
	}

	n, err := o.Write(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	if n != len(buf.Bytes()) {
		if err != nil {
			log.Fatal("could not write all bytes")
		}
	}
}

func buildLookup(dir, pkg string) (map[string]Export, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	f := ast.MergePackageFiles(
		pkgs[pkg],
		ast.FilterFuncDuplicates|
			ast.FilterImportDuplicates|
			ast.FilterUnassociatedComments,
	)

	lookup := make(map[string]Export)
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if isPublic(d.Name.Name) && d.Recv == nil {
				lookup[d.Name.Name] = Export{
					token.VAR,
					pkg,
					d.Name.Name,
					d.Doc,
				}
			}
		case *ast.GenDecl:
			if d.Tok == token.IMPORT {
				continue
			}

			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.ImportSpec:
					// pass
				case *ast.TypeSpec:
					doc := s.Doc
					if doc == nil {
						doc = s.Comment
					}
					if doc == nil {
						doc = d.Doc
					}

					if isPublic(s.Name.Name) && d.Doc != nil {
						lookup[s.Name.Name] = Export{
							d.Tok,
							pkg,
							s.Name.Name,
							doc,
						}
					}
				case *ast.ValueSpec:
					doc := s.Doc
					if doc == nil {
						doc = s.Comment
					}
					if doc == nil {
						doc = d.Doc
					}

					for _, name := range s.Names {
						if isPublic(name.Name) {
							lookup[name.Name] = Export{
								d.Tok,
								pkg,
								name.Name,
								doc,
							}
						}
					}
				default:
					log.Fatalf("unknown spec type %T", s)
				}
			}
		default:
			log.Fatalf("unknown declaration type %T", decl)
		}
	}

	return lookup, nil
}

func isPublic(name string) bool {
	return len(name) > 0 &&
		name[:1] != "_" &&
		strings.ToUpper(name[:1]) == name[:1]
}

func quicksort[T any, O constraints.Ordered](s []T, key func(T) O) {
	lo := 0
	hi := len(s) - 1
	qsort(s, key, lo, hi)
}

func qsort[T any, O constraints.Ordered](s []T, key func(T) O, lo, hi int) {
	p := partition(s, key, lo, hi)
	if p > 0 {
		qsort(s, key, lo, p-1)
	}
	if p < hi {
		qsort(s, key, p+1, hi)
	}
}

func partition[T any, O constraints.Ordered](
	s []T,
	key func(T) O,
	lo,
	hi int,
) int {
	pivot := key(s[hi])
	i := lo - 1
	for j := lo; j < hi; j++ {
		if key(s[j]) < pivot {
			i++
			s[i], s[j] = s[j], s[i]
		}
	}
	i++
	s[i], s[hi] = s[hi], s[i]
	return i
}
