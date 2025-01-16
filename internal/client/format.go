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

package gel

//go:generate go run golang.org/x/tools/cmd/stringer@v0.25.0 -type Format

// Format is the query response format.
type Format uint8

// IO Formats
const (
	Binary       Format = 0x62
	JSON         Format = 0x6a
	JSONElements Format = 0x4a
	Null         Format = 0x6e
)
