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

	"github.com/edgedb/edgedb-go/internal"
	edgedb "github.com/edgedb/edgedb-go/internal/client"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	"github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

var (
	errQueryNum = errors.New(
		"numbered query arguments detected, use named arguments instead",
	)
)

type queryConfig struct {
	name    string
	method  string
	file    string
	imports []string
	structs []*goStruct
	sTypes  *goStruct
	rTypes  []goType
}

type queryConfigV1 struct{}

type queryConfigV2 struct{}

type querySetup interface {
	setup(ctx context.Context, cmd, qryFile, outFile string,
		mixedCaps bool, c *edgedb.Client,
	) (*queryConfig, error)
}

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

	v, err := edgedb.ProtocolVersion(ctx, c)
	if err != nil {
		log.Fatalf("error determining the protocol version: %s", err)
	}

	var qs querySetup

	if v.GTE(internal.ProtocolVersion{Major: 2, Minor: 0}) {
		qs = &queryConfigV2{}
	} else {
		qs = &queryConfigV1{}
	}

	q, err := qs.setup(ctx, string(queryBytes), qryFile, outFile, mixedCaps, c)
	if err != nil {
		log.Fatalf("failed to setup query: %s", err)
	}

	return &Query{
		imports:             q.imports,
		QueryFile:           q.file,
		QueryName:           q.name,
		CMDVarName:          cmdVarName(qryFile),
		ResultTypes:         q.structs,
		SignatureReturnType: q.rTypes[0].Reference(),
		SignatureArgs:       q.sTypes.Fields,
		Method:              q.method,
	}, nil
}

func (r *queryConfigV1) setup(ctx context.Context, cmd, qryFile,
	outFile string, mixedCaps bool, c *edgedb.Client,
) (*queryConfig, error) {
	description, err := edgedb.Describe(ctx, c, cmd)

	if err != nil {
		return nil, fmt.Errorf("error introspecting query %q: %s", qryFile,
			err)
	}

	if isNumberedArgs(description.In) {
		return nil, errQueryNum
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

	return &queryConfig{
		qryName,
		m,
		qryFile,
		imports,
		rStructs,
		sTypes,
		rTypes,
	}, nil
}

func (r *queryConfigV2) setup(ctx context.Context, cmd, qryFile,
	outFile string, mixedCaps bool, c *edgedb.Client,
) (*queryConfig, error) {
	description, err := edgedb.DescribeV2(ctx, c, cmd)

	if err != nil {
		return nil, fmt.Errorf("error introspecting query %q: %s", qryFile,
			err)
	}

	if isNumberedArgsV2(&description.In) {
		return nil, errQueryNum
	}

	qryName := queryName(qryFile)
	rTypes, imports, err := resultTypesV2(qryName, description, mixedCaps)
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

	sTypes, i, err := signatureTypesV2(description, mixedCaps)
	if err != nil {
		log.Fatal(err)
	}
	imports = append(imports, i...)

	qryFile, err = queryFile(outFile, qryFile)
	if err != nil {
		log.Fatal(err)
	}

	m, err := methodV2(description)
	if err != nil {
		log.Fatal(err)
	}

	return &queryConfig{
		qryName,
		m,
		qryFile,
		imports,
		rStructs,
		sTypes,
		rTypes,
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

func signatureTypesV2(
	description *edgedb.CommandDescriptionV2,
	mixedCaps bool,
) (*goStruct, []string, error) {
	types, imports, err := generateTypeV2(&description.In,
		true, nil, mixedCaps)
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

func resultTypesV2(
	qryName string,
	description *edgedb.CommandDescriptionV2,
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
		outDesc = descriptor.V2{
			Type: descriptor.Set,
			ID:   id,
			Fields: []*descriptor.FieldV2{{
				Desc: description.Out,
			}},
		}
	case edgedb.One:
		required = true
	}

	return generateTypeV2(
		&outDesc,
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

func methodV2(description *edgedb.CommandDescriptionV2) (string, error) {
	switch description.Card {
	case edgedb.AtMostOne, edgedb.One:
		return "QuerySingle", nil
	case edgedb.NoResult, edgedb.Many, edgedb.AtLeastOne:
		return "Query", nil
	default:
		return "", errors.New("unreachable 20135")
	}
}
