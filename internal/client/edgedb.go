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
	"log"
	"time"

	"github.com/geldata/gel-go/internal"
	"github.com/geldata/gel-go/internal/buff"
	"github.com/geldata/gel-go/internal/cache"
	"github.com/geldata/gel-go/internal/codecs"
	"github.com/geldata/gel-go/internal/snc"
	"github.com/geldata/gel-go/internal/soc"
)

type cacheCollection struct {
	serverSettings    *snc.ServerSettings
	typeIDCache       *cache.Cache
	inCodecCache      *cache.Cache
	outCodecCache     *cache.Cache
	capabilitiesCache *cache.Cache // nolint:structcheck
}

type protocolConnection struct {
	soc                 *autoClosingSocket
	writeMemory         [1024]byte
	acquireReaderSignal chan struct{}
	readerChan          chan *buff.Reader

	protocolVersion internal.ProtocolVersion
	cacheCollection

	systemConfig systemConfig
	stateCodec   codecs.Encoder
}

// connectWithTimeout makes a single attempt to connect to `addr`.
func connectWithTimeout(
	ctx context.Context,
	cfg *connConfig,
	caches cacheCollection,
) (*protocolConnection, error) {
	socket, err := connectAutoClosingSocket(ctx, cfg)
	if err != nil {
		return nil, err
	}

	deadline, _ := ctx.Deadline()
	err = socket.SetDeadline(deadline)
	if err != nil {
		return nil, err
	}

	conn := &protocolConnection{
		soc:                 socket,
		acquireReaderSignal: make(chan struct{}, 1),
		readerChan:          make(chan *buff.Reader, 1),
		cacheCollection:     caches,
	}

	toBeDeserialized := make(chan *soc.Data, 2)
	go soc.Read(socket, soc.NewMemPool(4, 256*1024), toBeDeserialized)
	r := buff.NewReader(toBeDeserialized)

	err = conn.connect(r, cfg)
	if err != nil {
		return nil, err
	}

	err = socket.SetDeadline(time.Time{})
	if err != nil {
		return nil, err
	}

	return conn, conn.releaseReader(r)
}

func (c *protocolConnection) acquireReader(
	ctx context.Context,
) (*buff.Reader, error) {
	if c.isClosed() {
		return nil, &clientConnectionClosedError{}
	}

	c.acquireReaderSignal <- struct{}{}
	select {
	case r := <-c.readerChan:
		if r.Err != nil {
			return nil, &clientConnectionClosedError{err: r.Err}
		}
		if c.soc.Closed() {
			return nil, &clientConnectionClosedError{}
		}
		return r, nil
	case <-ctx.Done():
		return nil, wrapNetError(ctx.Err())
	}
}

func (c *protocolConnection) releaseReader(r *buff.Reader) error {
	if c.isClosed() {
		return &clientConnectionClosedError{}
	}

	if err := c.soc.SetDeadline(time.Time{}); err != nil {
		return err
	}

	go func() {
		for r.Next(c.acquireReaderSignal) {
			switch Message(r.MsgType) {
			case ErrorResponse:
				err := decodeErrorResponseMsg(r, "")
				log.Println("from background:", err)
				r.Err = wrapAll(r.Err, err)
			default:
				if e := c.fallThrough(r); e != nil {
					log.Println("fall through in background:", e)
					r.Err = wrapAll(r.Err, e)
					err := c.soc.Close()
					if err != nil {
						log.Println("error closing socket:", err)
					}
					c.readerChan <- r
					return
				}
			}
		}

		if r.Err != nil {
			err := c.soc.Close()
			if err != nil {
				log.Println("error closing socket:", err)
			}
		}
		c.readerChan <- r
	}()

	return nil
}

// Close the db connection
func (c *protocolConnection) close() error {
	if c.soc == nil {
		return &interfaceError{msg: "connection closed more than once"}
	}

	_, err := c.acquireReader(context.Background())
	if err != nil {
		return err
	}

	err = c.terminate()
	if err != nil {
		return err
	}

	return c.soc.Close()
}

func (c *protocolConnection) isClosed() bool {
	if c.soc == nil || c.soc.Closed() {
		return true
	}

	return false
}

func (c *protocolConnection) scriptFlow(ctx context.Context, q *query) error {
	if q.lang == SQL && c.protocolVersion.LT(protocolVersion3p0) {
		return &unsupportedFeatureError{
			msg: "the server does not support SQL queries, " +
				"upgrade to 6.0 or newer",
		}
	}

	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	deadline, _ := ctx.Deadline()
	err = c.soc.SetDeadline(deadline)
	if err != nil {
		return err
	}

	switch {
	case c.protocolVersion.GTE(protocolVersion2p0):
		err = c.execGranularFlow2pX(r, q)
	case c.protocolVersion.GTE(protocolVersion1p0):
		err = c.execGranularFlow1pX(r, q)
	default:
		err = c.execScriptFlow(r, q)
	}

	return firstError(err, c.releaseReader(r))
}

func (c *protocolConnection) granularFlow(
	ctx context.Context,
	q *query,
) error {
	if q.lang == SQL && c.protocolVersion.LT(protocolVersion3p0) {
		return &unsupportedFeatureError{
			msg: "the server does not support SQL queries, " +
				"upgrade to 6.0 or newer",
		}
	}

	r, err := c.acquireReader(ctx)
	if err != nil {
		return err
	}

	deadline, _ := ctx.Deadline()
	err = c.soc.SetDeadline(deadline)
	if err != nil {
		return err
	}

	switch {
	case c.protocolVersion.GTE(protocolVersion2p0):
		err = c.execGranularFlow2pX(r, q)
	case c.protocolVersion.GTE(protocolVersion1p0):
		err = c.execGranularFlow1pX(r, q)
	default:
		err = c.execGranularFlow0pX(r, q)
	}

	return firstError(err, c.releaseReader(r))
}
