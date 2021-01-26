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

package codecs

import "fmt"

// Path is used in error messages
// to show what field in a nested data structure caused the error.
type Path string

// AddField adds a field name to the path.
func (p Path) AddField(name string) Path {
	return Path(fmt.Sprintf("%v.%v", p, name))
}

// AddIndex adds an index to the path.
func (p Path) AddIndex(index int) Path {
	return Path(fmt.Sprintf("%v[%v]", p, index))
}
