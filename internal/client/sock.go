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
	"crypto/tls"
	"net"
	"sync"
	"time"
)

func connectAutoClosingSocket(
	ctx context.Context,
	cfg *connConfig,
) (*autoClosingSocket, error) {
	var cancel context.CancelFunc
	if cfg.connectTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cfg.connectTimeout)
		defer cancel()
	}

	conn, err := connectTLS(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &autoClosingSocket{conn: conn}, nil
}

func connectTLS(
	ctx context.Context,
	cfg *connConfig,
) (net.Conn, error) {
	tlsConfig, err := cfg.tlsConfig()
	if err != nil {
		return nil, err
	}

	d := tls.Dialer{Config: tlsConfig}
	conn, err := d.DialContext(ctx, cfg.addr.network, cfg.addr.address)
	if err != nil {
		return nil, wrapNetError(err)
	}

	protocol := conn.(*tls.Conn).ConnectionState().NegotiatedProtocol
	if protocol != "edgedb-binary" {
		_ = conn.Close()
		return nil, &clientConnectionFailedError{
			msg: "The server doesn't support the edgedb-binary protocol.",
		}
	}

	return conn, nil
}

// autoClosingSocket closes itself on network errors and future read/write
// operations fail immediately with an error.
type autoClosingSocket struct {
	conn     net.Conn
	isClosed bool
	mu       sync.Mutex
}

func (s *autoClosingSocket) Closed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isClosed
}

func (s *autoClosingSocket) Close() error {
	var err error
	s.mu.Lock()
	if !s.isClosed {
		s.isClosed = true
		err = s.conn.Close()
	}
	s.mu.Unlock()
	return wrapNetError(err)
}

func (s *autoClosingSocket) Read(p []byte) (int, error) {
	n, err := s.conn.Read(p)
	if err != nil {
		_ = s.Close()
		err = wrapNetError(err)
	}

	return n, err
}

func (s *autoClosingSocket) Write(p []byte) (int, error) {
	n, err := s.conn.Write(p)
	if err != nil {
		_ = s.Close()
		err = wrapNetError(err)
	}

	return n, err
}

func (s *autoClosingSocket) WriteAll(p []byte) error {
	for len(p) > 0 {
		n, err := s.Write(p)
		if err != nil {
			return err
		}
		p = p[n:]
	}

	return nil
}

func (s *autoClosingSocket) SetDeadline(t time.Time) error {
	err := s.conn.SetDeadline(t)
	if err != nil {
		_ = s.Close()
		err = wrapNetError(err)
	}

	return err
}
