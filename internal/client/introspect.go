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

package edgedb

import (
	"context"
)

// InstrospectionClient is an client with methods for introspecting.
type InstrospectionClient struct {
	*Client
}

// Describe returns CommandDescription for the provided cmd.
func (c *InstrospectionClient) Describe(
	ctx context.Context,
	cmd string,
) (*CommandDescription, error) {
	conn, err := c.acquire(ctx)
	if err != nil {
		return nil, err
	}

	q := &query{
		method:       "Query",
		cmd:          cmd,
		fmt:          Binary,
		expCard:      Many,
		capabilities: userCapabilities,
	}

	r, err := conn.conn.acquireReader(ctx)
	if err != nil {
		return nil, err
	}

	deadline, _ := ctx.Deadline()
	err = conn.conn.soc.SetDeadline(deadline)
	if err != nil {
		return nil, err
	}

	d, err := conn.conn.parse1pX(r, q)
	if err != nil {
		return nil, err
	}

	return d, nil
}
