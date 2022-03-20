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

package serverSettings

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/edgedb/edgedb-go/internal"
	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	"github.com/edgedb/edgedb-go/internal/edgedbtypes"
)

type systemConfig struct {
	ID                 edgedbtypes.UUID     `edgedb:"id"`
	SessionIdleTimeout edgedbtypes.Duration `edgedb:"session_idle_timeout"`
}

func parseSystemConfig(
	b []byte,
	version internal.ProtocolVersion,
) (*systemConfig, error) {
	if len(b) < 4+16+8+16+4 {
		return nil, fmt.Errorf("too few bytes")
	}
	r := buff.SimpleReader(b)
	u := r.PopSlice(r.PopUint32())
	u.PopUUID()
	dsc, err := descriptor.Pop(u, version)
	if err != nil {
		return nil, err
	}

	var cfg systemConfig
	typ := reflect.TypeOf(cfg)
	dec, err := codecs.BuildDecoder(dsc, typ, "system_config")
	if err != nil {
		return nil, err
	}

	err = dec.Decode(r.PopSlice(r.PopUint32()), unsafe.Pointer(&cfg))
	if err != nil {
		return nil, err
	}

	if len(r.Buf) != 0 {
		return nil, fmt.Errorf(
			"%v bytes left in buffer", len(r.Buf))
	}

	return &cfg, nil
}
