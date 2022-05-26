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
	"fmt"
	"reflect"

	"github.com/edgedb/edgedb-go/internal/buff"
	"github.com/edgedb/edgedb-go/internal/cardinality"
	"github.com/edgedb/edgedb-go/internal/codecs"
	"github.com/edgedb/edgedb-go/internal/descriptor"
	"github.com/edgedb/edgedb-go/internal/header"
	"github.com/edgedb/edgedb-go/internal/message"
)

func (c *protocolConnection) execGranularFlow1pX(
	r *buff.Reader,
	q *gfQuery,
) error {
	ids, ok := c.getCachedTypeIDs(q)
	if !ok {
		return c.pesimistic1pX(r, q)
	}

	cdcs, err := c.codecsFromIDs(ids, q)
	if err != nil {
		return err
	} else if cdcs == nil {
		return c.pesimistic1pX(r, q)
	}

	return c.optimistic1pX(r, q, cdcs)
}

func (c *protocolConnection) pesimistic1pX(r *buff.Reader, q *gfQuery) error {
	desc, err := c.parse1pX(r, q)
	if err != nil {
		return err
	}

	cdcs, err := c.codecsFromDescriptors(q, desc)
	if err != nil {
		return err
	}

	_, err = c.execute1pX(r, q, cdcs)
	return err
}

func (c *protocolConnection) optimistic1pX(
	r *buff.Reader,
	q *gfQuery,
	cdcs *codecPair,
) error {
	// When descriptors are returned the codec ids sent didn't match the
	// server's.  The codecs should be rebuilt with the new descriptors and the
	// execution retried.
	descs, err := c.execute1pX(r, q, cdcs)
	switch {
	case err == errZeroResults && descs != nil:
		goto Retry
	case err != nil:
		return err
	case descs != nil:
		// todo: correct error type
		return fmt.Errorf("unreachable")
	default:
		return nil
	}

Retry:
	cdcs, err = c.codecsFromDescriptors(q, descs)
	if err != nil {
		return err
	}

	descs, err = c.execute1pX(r, q, cdcs)
	if descs != nil {
		// todo: correct error type
		return fmt.Errorf("unexpected")
	}

	return err
}

func (c *protocolConnection) parse1pX(
	r *buff.Reader,
	q *gfQuery,
) (*descPair, error) {
	headers := copyHeaders(q.headers)
	headers[header.ExplicitObjectIDs] = []byte("true")

	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.Parse)
	writeHeaders(w, headers)
	w.PushUint8(q.fmt)
	w.PushUint8(q.expCard)
	w.PushString(q.cmd)
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := c.soc.WriteAll(w.Unwrap()); e != nil {
		return nil, &clientConnectionClosedError{err: e}
	}

	var (
		err  error
		desc *descPair
	)

	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.ParseComplete:
			var e error
			desc, e = c.decodeParseCompleteMsg1pX(r, q)
			err = wrapAll(err, e)
		case message.ReadyForCommand:
			decodeReadyForCommandMsg(r)
			done.Signal()
		case message.ErrorResponse:
			err = wrapAll(err, decodeErrorResponseMsg(r, q.cmd))
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return nil, e
			}
		}
	}

	if r.Err != nil || err != nil {
		return nil, wrapAll(r.Err, err)
	}

	return desc, nil
}

func (c *protocolConnection) decodeParseCompleteMsg1pX(
	r *buff.Reader,
	q *gfQuery,
) (*descPair, error) {
	c.cacheCapabilities(q, decodeHeaders(r))
	card := r.PopUint8()

	var (
		err   error
		descs descPair
	)

	id := r.PopUUID()
	descs.in, err = descriptor.Pop(
		r.PopSlice(r.PopUint32()),
		c.protocolVersion,
	)
	if err != nil {
		return nil, err
	}
	if descs.in.ID != id {
		return nil, &clientError{msg: fmt.Sprintf(
			"unexpected in descriptor id: %v",
			descs.in.ID,
		)}
	}

	id = r.PopUUID()
	descs.out, err = descriptor.Pop(
		r.PopSlice(r.PopUint32()),
		c.protocolVersion,
	)
	if err != nil {
		return nil, err
	}
	if descs.out.ID != id {
		return nil, &clientError{msg: fmt.Sprintf(
			"unexpected out descriptor id: %v",
			descs.out.ID,
		)}
	}

	if q.expCard == cardinality.AtMostOne && card == cardinality.Many {
		return nil, &resultCardinalityMismatchError{msg: fmt.Sprintf(
			"the query has cardinality %v "+
				"which does not match the expected cardinality %v",
			cardinality.ToStr[card],
			cardinality.ToStr[q.expCard],
		)}
	}

	c.cacheTypeIDs(q, idPair{in: descs.in.ID, out: descs.out.ID})
	descCache.Put(descs.in.ID, descs.in)
	descCache.Put(descs.out.ID, descs.out)
	return &descs, nil
}

func (c *protocolConnection) execute1pX(
	r *buff.Reader,
	q *gfQuery,
	cdcs *codecPair,
) (*descPair, error) {
	headers := copyHeaders(q.headers)
	headers[header.ExplicitObjectIDs] = []byte("true")

	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.Execute)
	writeHeaders(w, headers)
	w.PushUint8(q.fmt)
	w.PushUint8(q.expCard)
	w.PushString(q.cmd)
	w.PushUUID(cdcs.in.DescriptorID())
	w.PushUUID(cdcs.out.DescriptorID())
	if e := cdcs.in.Encode(w, q.args, codecs.Path("args"), true); e != nil {
		return nil, &invalidArgumentError{msg: e.Error()}
	}
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := c.soc.WriteAll(w.Unwrap()); e != nil {
		return nil, &clientConnectionClosedError{err: e}
	}

	tmp := q.out
	err := error(nil)
	if q.expCard == cardinality.AtMostOne {
		err = errZeroResults
	}
	done := buff.NewSignal()

	var descs *descPair
	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.Data:
			val, ok, e := decodeDataMsg(r, q, cdcs)
			if e != nil {
				if err == errZeroResults {
					err = e
				} else {
					err = wrapAll(err, e)
				}
			}
			if ok {
				tmp = reflect.Append(tmp, val)
			}

			if err == errZeroResults {
				err = nil
			}
		case message.CommandComplete:
			decodeCommandCompleteMsg(r)
		case message.CommandDataDescription:
			fallthrough
		case message.ParseComplete:
			var e error
			descs, e = c.decodeParseCompleteMsg1pX(r, q)
			err = wrapAll(err, e)
		case message.ReadyForCommand:
			decodeReadyForCommandMsg(r)
			done.Signal()
		case message.ErrorResponse:
			if err == errZeroResults {
				err = nil
			}

			err = wrapAll(err, decodeErrorResponseMsg(r, q.cmd))
		default:
			if e := c.fallThrough(r); e != nil {
				// the connection will not be usable after this x_x
				return nil, e
			}
		}
	}

	if r.Err != nil {
		return nil, r.Err
	}

	if !q.flat() {
		q.out.Set(tmp)
	}

	return descs, err
}