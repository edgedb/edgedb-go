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
	"github.com/edgedb/edgedb-go/internal/format"
	"github.com/edgedb/edgedb-go/internal/message"
	"github.com/edgedb/edgedb-go/internal/state"
)

func (c *protocolConnection) execGranularFlow1pX(
	r *buff.Reader,
	q *query,
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

	return c.execute1pX(r, q, cdcs)
}

func (c *protocolConnection) pesimistic1pX(r *buff.Reader, q *query) error {
	desc, err := c.parse1pX(r, q)
	if err != nil {
		return err
	}

	cdcs, err := c.codecsFromDescriptors1pX(q, desc)
	if err != nil {
		return err
	}

	return c.execute1pX(r, q, cdcs)
}

func (c *protocolConnection) parse1pX(
	r *buff.Reader,
	q *query,
) (*descPair, error) {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.Parse)
	w.PushUint16(0) // no headers
	w.PushUint64(q.capabilities)
	w.PushUint64(0) // no compilation_flags
	w.PushUint64(0) // no implicit limit
	w.PushUint8(q.fmt)
	w.PushUint8(q.expCard)
	w.PushString(q.cmd)

	w.PushUUID(c.stateCodec.DescriptorID())
	if e := c.stateCodec.Encode(w, codecs.Path("state"), q.state); e != nil {
		return nil, &binaryProtocolError{err: fmt.Errorf(
			"invalid connection state: %w", e)}
	}
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
		case message.StateDataDescription:
			if e := c.decodeStateDataDescription(r); e != nil {
				err = wrapAll(err, e)
			}
		case message.CommandDataDescription:
			var e error
			desc, e = c.decodeCommandDataDescriptionMsg1pX(r, q)
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

func (c *protocolConnection) decodeCommandDataDescriptionMsg1pX(
	r *buff.Reader,
	q *query,
) (*descPair, error) {
	discardHeaders(r)
	c.cacheCapabilities1pX(q, r.PopUint64())
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
	} else if descs.in.ID != id {
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
	} else if descs.out.ID != id {
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
	q *query,
	cdcs *codecPair,
) error {
	w := buff.NewWriter(c.writeMemory[:0])
	w.BeginMessage(message.Execute)
	w.PushUint16(0) // no headers
	w.PushUint64(q.capabilities)
	w.PushUint64(0) // no compilation_flags
	w.PushUint64(0) // no implicit limit
	w.PushUint8(q.fmt)
	w.PushUint8(q.expCard)
	w.PushString(q.cmd)

	w.PushUUID(c.stateCodec.DescriptorID())
	if e := c.stateCodec.Encode(w, codecs.Path("state"), q.state); e != nil {
		return &binaryProtocolError{err: fmt.Errorf(
			"invalid connection state: %w", e)}
	}

	w.PushUUID(cdcs.in.DescriptorID())
	w.PushUUID(cdcs.out.DescriptorID())
	if e := cdcs.in.Encode(w, q.args, codecs.Path("args"), true); e != nil {
		return &invalidArgumentError{msg: e.Error()}
	}
	w.EndMessage()

	w.BeginMessage(message.Sync)
	w.EndMessage()

	if e := c.soc.WriteAll(w.Unwrap()); e != nil {
		return &clientConnectionClosedError{err: e}
	}

	tmp := q.out
	err := error(nil)
	if q.expCard == cardinality.AtMostOne {
		err = errZeroResults
	}
	done := buff.NewSignal()

	for r.Next(done.Chan) {
		switch r.MsgType {
		case message.StateDataDescription:
			if e := c.decodeStateDataDescription(r); e != nil {
				err = wrapAll(err, e)
			}
		case message.CommandDataDescription:
			descs, e := c.decodeCommandDataDescriptionMsg1pX(r, q)
			err = wrapAll(err, e)
			cdcs, e = c.codecsFromDescriptors1pX(q, descs)
			err = wrapAll(err, e)
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
			if e := c.decodeCommandCompleteMsg1pX(q, r); e != nil {
				err = wrapAll(err, e)
			}
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
				return e
			}
		}
	}

	if r.Err != nil {
		return r.Err
	}

	if !q.flat() && q.fmt != format.Null {
		q.out.Set(tmp)
	}

	return err
}

func (c *protocolConnection) codecsFromDescriptors1pX(
	q *query,
	descs *descPair,
) (*codecPair, error) {
	var cdcs codecPair
	var err error
	cdcs.in, err = codecs.BuildEncoder(descs.in, c.protocolVersion)
	if err != nil {
		return nil, &invalidArgumentError{msg: err.Error()}
	}

	if q.fmt == format.JSON {
		cdcs.out = codecs.JSONBytes
	} else {
		var path codecs.Path
		if q.fmt == format.Null {
			// There is no outType value for Null output format queries.
			path = "null"
		} else {
			path = codecs.Path(q.outType.String())
		}

		cdcs.out, err = codecs.BuildDecoder(descs.out, q.outType, path)
		if err != nil {
			err = fmt.Errorf(
				"the \"out\" argument does not match query schema: %v",
				err,
			)
			return nil, &invalidArgumentError{msg: err.Error()}
		}
	}

	c.inCodecCache.Put(cdcs.in.DescriptorID(), cdcs.in)
	c.outCodecCache.Put(
		codecKey{ID: cdcs.out.DescriptorID(), Type: q.outType},
		cdcs.out,
	)

	return &cdcs, nil
}

func (c *protocolConnection) decodeCommandCompleteMsg1pX(
	q *query,
	r *buff.Reader,
) error {
	discardHeaders(r)
	c.cacheCapabilities1pX(q, r.PopUint64())
	r.Discard(int(r.PopUint32())) // discard command status
	if r.PopUUID() == descriptor.IDZero {
		// empty state data
		r.Discard(4)
		return nil
	}

	r.Discard(int(r.PopUint32())) // state data
	return nil
}

func (c *protocolConnection) decodeStateDataDescription(r *buff.Reader) error {
	id := r.PopUUID()
	desc, err := descriptor.Pop(
		r.PopSlice(r.PopUint32()),
		c.protocolVersion,
	)
	if err != nil {
		return &binaryProtocolError{err: fmt.Errorf(
			"decoding ParameterStatus state_description: %w", err)}
	} else if desc.ID != id {
		return &binaryProtocolError{err: fmt.Errorf(
			"state_description ids don't match: %v != %v", id, desc.ID)}
	}

	codec, err := state.BuildCodec(desc, codecs.Path("state"))
	if err != nil {
		return &binaryProtocolError{err: fmt.Errorf(
			"building decoder from ParameterStatus state_description: %w",
			err)}
	}

	c.stateCodec = codec
	return nil
}
