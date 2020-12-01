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
	"log"
	"net"

	"github.com/edgedb/edgedb-go/cache"
	"github.com/edgedb/edgedb-go/marshal"
	"github.com/edgedb/edgedb-go/protocol/cardinality"
	"github.com/edgedb/edgedb-go/protocol/format"
)

// todo add examples

type baseConn struct {
	conn           net.Conn
	buffer         [8192]byte
	typeIDCache    *cache.Cache
	inCodecCache   *cache.Cache
	outCodecCache  *cache.Cache
	serverSettings map[string]string
}

// ConnectOne establishes a connection to an EdgeDB server.
func ConnectOne(ctx context.Context, opts Options) (*Conn, error) { // nolint:gocritic,lll
	return ConnectOneDSN(ctx, "", opts)
}

// ConnectOneDSN establishes a connection to an EdgeDB server.
func ConnectOneDSN(
	ctx context.Context,
	dsn string,
	opts Options, // nolint:gocritic
) (*Conn, error) {
	conn := &baseConn{
		typeIDCache:   cache.New(1_000),
		inCodecCache:  cache.New(1_000),
		outCodecCache: cache.New(1_000),
	}

	config, err := parseConnectDSNAndArgs(dsn, &opts)
	if err != nil {
		return nil, err
	}

	if err := connectOne(ctx, config, conn); err != nil {
		return nil, err
	}

	return &Conn{*conn}, nil
}

// connectOne expectes a singleConn that has a nil net.Conn.
func connectOne(ctx context.Context, cfg *connConfig, conn *baseConn) error {
	var (
		d   net.Dialer
		err error
	)

	for _, addr := range cfg.addrs { // nolint:gocritic
		// todo do error values need to be checked?
		conn.conn, err = d.DialContext(ctx, addr.network, addr.address)
		if err != nil {
			log.Printf("while attempting connection %+v: %+v", addr, err)
			continue
		}

		err = conn.connect(ctx, cfg)
		if err != nil {
			_ = conn.conn.Close()
			log.Printf("while attempting connection %+v: %+v", addr, err)
			continue
		}

		return nil
	}

	conn.conn = nil
	return err
}

// Close the db connection
func (c *baseConn) close() error {
	return wrapAll(c.terminate(), c.conn.Close())
}

// Execute an EdgeQL command (or commands).
func (c *baseConn) Execute(ctx context.Context, cmd string) (err error) {
	return c.scriptFlow(ctx, c.conn, cmd)
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// ErrorZeroResults is returned.
func (c *baseConn) QueryOne(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) (err error) {
	val, err := marshal.ValueOf(out)
	if err != nil {
		return err
	}

	q := query{
		cmd:     cmd,
		fmt:     format.Binary,
		expCard: cardinality.One,
		args:    args,
	}

	err = c.granularFlow(ctx, val, q)
	if err != nil {
		return err
	}

	return nil
}

// Query runs a query and returns the results.
func (c *baseConn) Query(
	ctx context.Context,
	cmd string,
	out interface{},
	args ...interface{},
) error {
	val, err := marshal.ValueOfSlice(out)
	if err != nil {
		return err
	}

	q := query{
		cmd:     cmd,
		fmt:     format.Binary,
		expCard: cardinality.Many,
		args:    args,
	}

	err = c.granularFlow(ctx, val, q)
	if err != nil {
		return err
	}

	return nil
}

// QueryJSON runs a query and return the results as JSON.
func (c *baseConn) QueryJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	val, err := marshal.ValueOf(out)
	if err != nil {
		return err
	}

	q := query{
		cmd:     cmd,
		fmt:     format.JSON,
		expCard: cardinality.Many,
		args:    args,
	}

	err = c.granularFlow(ctx, val, q)
	if err != nil {
		return err
	}

	return nil
}

// QueryOneJSON runs a singleton-returning query
// and return its element in JSON.
// If the query executes successfully but doesn't return a result
// []byte{}, ErrorZeroResults is returned.
func (c *baseConn) QueryOneJSON(
	ctx context.Context,
	cmd string,
	out *[]byte,
	args ...interface{},
) error {
	val, err := marshal.ValueOf(out)
	if err != nil {
		return err
	}

	q := query{
		cmd:     cmd,
		fmt:     format.JSON,
		expCard: cardinality.One,
		args:    args,
	}

	err = c.granularFlow(ctx, val, q)
	if err != nil {
		return err
	}

	if len(*out) == 0 {
		return ErrZeroResults
	}

	return nil
}

func (c *baseConn) writeAndRead(
	ctx context.Context,
	buf *[]byte,
) (err error) {
	// todo move set deadline up to query method.
	deadline, _ := ctx.Deadline()
	err = c.conn.SetDeadline(deadline)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(*buf)
	if err != nil {
		return err
	}

	// expand slice length to full capacity
	*buf = (*buf)[:cap(*buf)]

	n, err := c.conn.Read(*buf)
	*buf = (*buf)[:n]

	if n < cap(*buf) {
		return err
	}

	n = 1024 // todo evaluate temporary buffer size
	tmp := make([]byte, n)
	for n == 1024 {
		n, err = c.conn.Read(tmp)
		*buf = append(*buf, tmp[:n]...)
	}

	return err
}
