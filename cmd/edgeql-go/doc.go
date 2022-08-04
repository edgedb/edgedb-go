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

// edgeql-go is a tool to generate go functions from edgeql queries. Given the
// name of one or more edgeql files edgeql-go will create a new self-contained
// Go source with a function for each edgeql query.
//
// # Install
//
//	go install github.com/edgedb/edgedb-go/cmd/edgeql-go@latest
//
// See this link for [pinning tool dependencies].
//
// # Usage
//
// Typically this process would be run using go generate, like this:
//
//	//go:generate edgeql-go -file=myquery.edgeql
//
// The -file flag may be specified multiple times to include more than one
// query in the output file.
//
// The -out flag is the path to the go source file to be created. If omitted,
// the file is created next to the first edgeql file provided.
//
// The -json flag tells edgeql-go to return json from the query instead of go
// structs.
//
// [pinning tool dependencies]: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package main
