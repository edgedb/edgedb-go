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

import (
	"context"

	"github.com/geldata/gel-go/internal"
	"github.com/geldata/gel-go/internal/descriptor"
)

// CommandDescription is the information returned in the CommandDataDescription
// message
type CommandDescription struct {
	In   descriptor.Descriptor
	Out  descriptor.Descriptor
	Card Cardinality
}

// CommandDescriptionV2 is the information returned in the
// CommandDataDescription message
type CommandDescriptionV2 struct {
	In   descriptor.V2
	Out  descriptor.V2
	Card Cardinality
}

// Describe returns CommandDescription for the provided cmd.
func Describe(
	ctx context.Context,
	c *Client,
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

// DescribeV2 returns CommandDescription for the provided cmd.
func DescribeV2(
	ctx context.Context,
	c *Client,
	cmd string,
) (*CommandDescriptionV2, error) {
	conn, err := c.acquire(ctx)
	if err != nil {
		return nil, err
	}

	q := &query{
		method:       "Query",
		lang:         EdgeQL,
		cmd:          cmd,
		fmt:          Binary,
		expCard:      Many,
		capabilities: userCapabilities,
		parse:        true,
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

	d, err := conn.conn.parse2pX(r, q)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// ProtocolVersion returns the protocol version used by c.
func ProtocolVersion(
	ctx context.Context,
	c *Client,
) (internal.ProtocolVersion, error) {
	conn, err := c.acquire(ctx)
	if err != nil {
		return internal.ProtocolVersion{}, err
	}

	protocolVersion := conn.conn.protocolVersion
	err = c.release(conn, nil)
	if err != nil {
		return internal.ProtocolVersion{}, err
	}

	return protocolVersion, nil
}
