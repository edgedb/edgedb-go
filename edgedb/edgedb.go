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
	"errors"
	"net"
	"reflect"

	"github.com/edgedb/edgedb-go/edgedb/cache"
	"github.com/edgedb/edgedb-go/edgedb/marshal"
	"github.com/edgedb/edgedb-go/edgedb/protocol/cardinality"
	"github.com/edgedb/edgedb-go/edgedb/protocol/format"
)

// todo add examples

var (
	// todo should this be returned from Query() and QueryJSON()? :thinking:

	// ErrorZeroResults is returned when a query has no results.
	ErrorZeroResults = errors.New("zero results")
)

// Conn is a connection to an EdgeDB server.
type Conn struct {
	conn          net.Conn
	buffer        [8192]byte
	typeIDCache   *cache.Cache
	inCodecCache  *cache.Cache
	outCodecCache *cache.Cache
}

// Close the db connection
func (c *Conn) Close() error {
	return c.conn.Close()
}

// Execute an EdgeQL command (or commands).
func (c *Conn) Execute(ctx context.Context, query string) error {
	return c.scriptFlow(ctx, query)
}

// QueryOne runs a singleton-returning query and returns its element.
// If the query executes successfully but doesn't return a result
// ErrorZeroResults is returned.
func (c *Conn) QueryOne(
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
func (c *Conn) Query(
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
func (c *Conn) QueryJSON(
	ctx context.Context,
	cmd string,
	args ...interface{},
) ([]byte, error) {
	q := query{
		cmd:     cmd,
		fmt:     format.JSON,
		expCard: cardinality.Many,
		args:    args,
	}

	var result []byte
	val := reflect.ValueOf(&result).Elem()
	err := c.granularFlow(ctx, val, q)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// QueryOneJSON runs a singleton-returning query
// and return its element in JSON.
// If the query executes successfully but doesn't return a result
// []byte{}, ErrorZeroResults is returned.
func (c *Conn) QueryOneJSON(
	ctx context.Context,
	cmd string,
	args ...interface{},
) ([]byte, error) {
	q := query{
		cmd:     cmd,
		fmt:     format.JSON,
		expCard: cardinality.One,
		args:    args,
	}

	var result []byte
	val := reflect.ValueOf(&result).Elem()
	err := c.granularFlow(ctx, val, q)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrorZeroResults
	}

	return result, nil
}

// Connect establishes a connection to an EdgeDB server.
func Connect(ctx context.Context, opts Options) (*Conn, error) {
	var d net.Dialer
	c, err := d.DialContext(ctx, opts.network(), opts.address())
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		conn:          c,
		typeIDCache:   cache.New(1_000),
		inCodecCache:  cache.New(1_000),
		outCodecCache: cache.New(1_000),
	}

	err = conn.connect(ctx, &opts)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *Conn) writeAndRead(ctx context.Context, buf *[]byte) error {
	deadline, _ := ctx.Deadline()
	err := c.conn.SetDeadline(deadline)
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
