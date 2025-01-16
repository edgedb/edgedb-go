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

//go:generate go run golang.org/x/tools/cmd/stringer@v0.25.0 -type Cardinality

// Cardinality is the result cardinality for a command.
type Cardinality uint8

// Cardinalities
const (
	NoResult   Cardinality = 0x6e
	AtMostOne  Cardinality = 0x6f
	One        Cardinality = 0x41
	Many       Cardinality = 0x6d
	AtLeastOne Cardinality = 0x4d
)
