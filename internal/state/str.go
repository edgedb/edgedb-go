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

package state

import (
	"fmt"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

type strCodec struct {
	id edgedbtypes.UUID
}

func (c *strCodec) DescriptorID() edgedbtypes.UUID { return c.id }

func (c *strCodec) Decode(
	r *buff.Reader,
	path codecs.Path,
) (interface{}, error) {
	return string(r.Buf), nil
}

func (c *strCodec) Encode(
	w *buff.Writer,
	path codecs.Path,
	val interface{},
) error {
	in, ok := val.(string)
	if !ok {
		return fmt.Errorf("expected %v to be string got: %T", path, val)
	}

	w.PushString(in)
	return nil
}
