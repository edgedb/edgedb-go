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

// edgeql-go is a tool to generate go functions from edgeql queries. When run
// in an Gel project directory (or subdirectory) a *_edgeql.go source file
// will be generated for each *.edgeql file.  The generated go will have an
// edgeqlFileName and edgeqlFileNameJSON function with typed arguments and
// return value matching the query's arguments and result shape.
//
// # Install
//
//	go install github.com/geldata/gel-go/cmd/edgeql-go@latest
//
// See also [pinning tool dependencies].
//
// # Usage
//
// Typically this process would be run using [go generate] like this:
//
//	//go:generate edgeql-go -pubfuncs -pubtypes -mixedcaps
//
// For a complete list of options:
//
//	edgeql-go -help
//
// [pinning tool dependencies]: https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
// [go generate]: https://go.dev/blog/generate
package main
