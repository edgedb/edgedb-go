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

package main

import (
	"context"
	"crypto/rand"
	"errors"
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
	qryFile,
	outFile string,
	mixedCaps bool,
) (*Query, error) {
	var err error
	qryFile, err = filepath.Abs(qryFile)
	if err != nil {
		return nil, err
	}

	queryBytes, err := os.ReadFile(qryFile)
	if err != nil {
		log.Fatalf("error reading %q: %s", qryFile, err)
	}

	description, err := edgedb.Describe(ctx, c, string(queryBytes))
	if err != nil {
		log.Fatalf("error introspecting query %q: %s", qryFile, err)
	}

	if isNumberedArgs(description.In) {
		log.Fatalf(
			"numbered query arguments detected, use named arguments instead",
		)
	}

	qryName := queryName(qryFile)
	rTypes, imports, err := resultTypes(qryName, description, mixedCaps)
	if err != nil {
		log.Fatal(err)
	}
	var rStructs []*goStruct
	for _, typ := range rTypes {
		if t, ok := typ.(*goStruct); ok {
			t.QueryFuncName = qryName
			rStructs = append(rStructs, t)
		}
	}

	sTypes, i, err := signatureTypes(description, mixedCaps)
	if err != nil {
		log.Fatal(err)
	}
	imports = append(imports, i...)

	qryFile, err = queryFile(outFile, qryFile)
	if err != nil {
		log.Fatal(err)
	}

	m, err := method(description)
	if err != nil {
		log.Fatal(err)
	}

	return &Query{
		imports: imports,

		QueryFile:           qryFile,
		QueryName:           qryName,
		CMDVarName:          cmdVarName(qryFile),
		ResultTypes:         rStructs,
		SignatureReturnType: rTypes[0].Reference(),
		SignatureArgs:       sTypes.Fields,
		Method:              m,
	}, nil
}

func queryFile(outFile, queryFile string) (string, error) {
	return filepath.Rel(filepath.Dir(outFile), queryFile)
}

func cmdVarName(qryFile string) string {
	name := filepath.Base(qryFile)
	name = strings.TrimSuffix(name, ".edgeql")
	name = fmt.Sprintf("%s_cmd", name)
	return snakeToLowerMixedCase(name)
}

func queryName(qryFile string) string {
	name := filepath.Base(qryFile)
	name = strings.TrimSuffix(name, ".edgeql")
	return snakeToLowerMixedCase(name)
}

func signatureTypes(
	description *edgedb.CommandDescription,
	mixedCaps bool,
) (*goStruct, []string, error) {
	types, imports, err := generateType(description.In, true, nil, mixedCaps)
	if err != nil {
		return &goStruct{}, nil, err
	}

	return types[0].(*goStruct), imports, nil
}

func resultTypes(
	qryName string,
	description *edgedb.CommandDescription,
	mixedCaps bool,
) ([]goType, []string, error) {
	outDesc := description.Out
	var required bool
	switch description.Card {
	case edgedb.Many, edgedb.AtLeastOne:
		id, err := randomID()
		if err != nil {
			return nil, nil, err
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

	return generateType(
		outDesc,
		required,
		[]string{qryName + "Result"},
		mixedCaps,
	)
}

func randomID() (edgedbtypes.UUID, error) {
	var id edgedbtypes.UUID
	_, err := rand.Read(id[:])
	return id, err
}

func method(description *edgedb.CommandDescription) (string, error) {
	switch description.Card {
	case edgedb.AtMostOne, edgedb.One:
		return "QuerySingle", nil
	case edgedb.NoResult, edgedb.Many, edgedb.AtLeastOne:
		return "Query", nil
	default:
		return "", errors.New("unreachable 20135")
	}
}
