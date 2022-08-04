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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	edgedb "github.com/edgedb/edgedb-go/internal/client"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	types "github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

func newQuery(
	ctx context.Context,
	client edgedb.InstrospectionClient,
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

	description, err := client.Describe(ctx, string(queryBytes))
	if err != nil {
		log.Fatalf("error introspecting query %q: %s", queryFile, err)
	}

	if isNumberedArgs(description.In) {
		log.Fatalf(
			"numbered query arguments detected, use named arguments instead",
		)
		// todo: maybe check that argument names are valid identifiers
	}

	return &Query{queryFile, outFile, description}, nil
}

// Query generates values for templates/query.template
type Query struct {
	queryFile   string
	outFile     string
	description *edgedb.CommandDescription
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
	return snakeToUpperCamelCase(name)
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

		name := snakeToLowerCamelCase(field.Name)
		fmt.Fprintf(&buf, "\n%s %s,", name, typ)
	}

	return buf.String(), nil
}

// ResultType returns the type that the query returns.
func (q *Query) ResultType() (string, error) {
	if *isJSON {
		return "[]byte", nil
	}

	outDesc := q.description.Out
	var required bool
	switch q.description.Card {
	case edgedb.Many, edgedb.AtLeastOne:
		required = true
		outDesc = descriptor.Descriptor{
			Type: descriptor.Set,
			ID:   types.UUID{},
			Fields: []*descriptor.Field{{
				Desc: q.description.Out,
			}},
		}
	case edgedb.One:
		required = true
	}

	return typeGen.getType(outDesc, required)
}

// ArgList returns then list of arguments to pass to the query method.
func (q *Query) ArgList() (string, error) {
	if len(q.description.In.Fields) == 0 {
		return "", nil
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "\nmap[string]interface{}{\n")
	for _, field := range q.description.In.Fields {
		name := snakeToLowerCamelCase(field.Name)
		fmt.Fprintf(&buf, "%q: %s,\n", name, name)
	}
	fmt.Fprintf(&buf, "},")

	return buf.String(), nil
}

// Method returns the edgedb.Client query method name.
func (q *Query) Method() string {
	switch q.description.Card {
	case edgedb.AtMostOne, edgedb.One:
		if *isJSON {
			return "QuerySingleJSON"
		}
		return "QuerySingle"
	case edgedb.NoResult, edgedb.Many, edgedb.AtLeastOne:
		if *isJSON {
			return "QueryJSON"
		}
		return "Query"
	default:
		panic("unreachable 20135")
	}
}
