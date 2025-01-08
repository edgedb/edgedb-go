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

package gel_test

import (
	"context"
	"testing"

	gel "github.com/edgedb/edgedb-go"
)

var (
	ctx    context.Context
	client *gel.Client
)

// [Link properties] are treated as fields in the linked to struct, and the @
// is omitted from the field's tag.
//
// [Link properties]: https://www.edgedb.com/docs/guides/link_properties
func Example_linkProperty() {
	var result []struct {
		Name    string `gel:"name"`
		Friends []struct {
			Name     string              `gel:"name"`
			Strength gel.OptionalFloat64 `gel:"strength"`
		} `gel:"friends"`
	}

	_ = client.Query(
		ctx,
		`select Person {
		name,
		friends: {
			name,
			@strength,
		}
	}`,
		&result,
	)
}

// TestNil makes this not a whole file example. // https://go.dev/blog/examples
func TestNil(t *testing.T) {}
