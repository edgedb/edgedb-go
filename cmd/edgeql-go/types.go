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

// Query is used in templates/Query.template
type Query struct {
	QueryFile           string
	QueryName           string
	CMDVarName          string
	ResultTypes         []*goStruct
	SignatureReturnType string
	SignatureArgs       []goStructField
	Method              string

	imports []string
}

type goType interface {
	// Reference is the name used to refer to the type.
	Reference() string
}

type goScalar struct {
	Name string
}

func (t *goScalar) Reference() string { return t.Name }

type goSlice struct {
	typ goType
}

func (t *goSlice) Reference() string { return "[]" + t.typ.Reference() }

type goStructField struct {
	EQLName string
	GoName  string
	Type    string
	Tag     string
}

type goStruct struct {
	Name          string
	QueryFuncName string
	Fields        []goStructField
}

func (t *goStruct) Reference() string { return t.Name }
