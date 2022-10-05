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
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	edgedb "github.com/edgedb/edgedb-go/internal/client"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	"github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

func newQuery(
	ctx context.Context,
	c *edgedb.Client,
	queryFile,
	outFile string,
) (*Query, error) {
	var err error
	queryFile, err = filepath.Abs(queryFile)
	if err != nil {
		return nil, err
	}

	queryBytes, err := os.ReadFile(queryFile)
	if err != nil {
		log.Fatalf("error reading %q: %s", queryFile, err)
	}

	description, err := edgedb.Describe(ctx, c, string(queryBytes))
	if err != nil {
		log.Fatalf("error introspecting query %q: %s", queryFile, err)
	}

	if isNumberedArgs(description.In) {
		log.Fatalf(
			"numbered query arguments detected, use named arguments instead",
		)
		// todo: maybe check that argument names are valid identifiers
	}

	rType, err := resultType(description)
	if err != nil {
		log.Fatal(err)
	}

	sTypes, err := signatureTypes(description)
	if err != nil {
		log.Fatal(err)
	}

	return &Query{queryFile, outFile, description, rType, sTypes}, nil
}

// Query generates values for templates/query.template
type Query struct {
	queryFile      string
	outFile        string
	description    *edgedb.CommandDescription
	resultType     Type
	signatureTypes []Type
}

// QueryFile returns the relative path from the go source file to the edgeql
// file.
func (q *Query) QueryFile() (string, error) {
	return filepath.Rel(filepath.Dir(q.outFile), q.queryFile)
}

// CMDVarName returns the name of the variable that embeds the edgeql file.
func (q *Query) CMDVarName() string {
	name := filepath.Base(q.queryFile)
	name = strings.TrimSuffix(name, ".edgeql")
	name = fmt.Sprintf("%s_cmd", name)
	return snakeToLowerCamelCase(name)
}

// QueryName returns the name of the function that will run the query.
func (q *Query) QueryName() string {
	name := filepath.Base(q.queryFile)
	name = strings.TrimSuffix(name, ".edgeql")
	return snakeToLowerCamelCase(name)
}

func signatureTypes(description *edgedb.CommandDescription) ([]Type, error) {
	types := make([]Type, len(description.In.Fields))

	for _, field := range description.In.Fields {
		typ, err := typeGen.getType(field.Desc, field.Required)
		if err != nil {
			return nil, err
		}
		types = append(types, typ)
	}

	return types, nil
}

// SignatureArgs returns the query arguments as they will appear  in the
// function signature.
func (q *Query) SignatureArgs() (string, error) {
	var buf bytes.Buffer
	for _, field := range q.description.In.Fields {
		typ, err := typeGen.getType(field.Desc, field.Required)
		if err != nil {
			return "", err
		}

		fmt.Fprintf(&buf, "\n%s %s,", field.Name, typ.definition)
	}

	return buf.String(), nil
}

func resultType(description *edgedb.CommandDescription) (Type, error) {
	outDesc := description.Out
	var required bool
	switch description.Card {
	case edgedb.Many, edgedb.AtLeastOne:
		id, err := randomID()
		if err != nil {
			return Type{}, err
		}

		required = true
		outDesc = descriptor.Descriptor{
			Type: descriptor.Set,
			ID:   id,
			Fields: []*descriptor.Field{{
				Desc: description.Out,
			}},
		}
	case edgedb.One:
		required = true
	}

	return typeGen.getType(outDesc, required)
}

func randomID() (edgedbtypes.UUID, error) {
	var id edgedbtypes.UUID
	_, err := rand.Read(id[:])
	return id, err
}

// ResultType returns the type declaration for the query result.
func (q *Query) ResultType() string {
	return q.resultType.definition
}

// ArgList returns then list of arguments to pass to the query method.
func (q *Query) ArgList() (string, error) {
	if len(q.description.In.Fields) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "\nmap[string]interface{}{\n")
	for _, field := range q.description.In.Fields {
		fmt.Fprintf(&buf, "%q: %s,\n", field.Name, field.Name)
	}
	fmt.Fprintf(&buf, "},")

	return buf.String(), nil
}

// Method returns the edgedb.Client query method name.
func (q *Query) Method() string {
	switch q.description.Card {
	case edgedb.AtMostOne, edgedb.One:
		return "QuerySingle"
	case edgedb.NoResult, edgedb.Many, edgedb.AtLeastOne:
		return "Query"
	default:
		panic("unreachable 20135")
	}
}

// Imports returns extra packages that need to be imported.
func (q *Query) Imports() []string {
	imports := q.resultType.imports
	for i := 0; i < len(q.signatureTypes); i++ {
		imports = append(imports, q.signatureTypes[i].imports...)
	}

	unique := make(map[string]struct{})
	for _, name := range imports {
		unique[name] = struct{}{}
	}

	result := make([]string, len(unique))
	result = result[:0]
	for name := range unique {
		result = append(result, name)
	}

	return result
}
