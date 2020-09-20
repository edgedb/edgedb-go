package payload

import (
	"encoding/binary"
	"fmt"
)

// Payload a FIFO queue of Messages
type Payload struct {
	buf []byte
}

// Make a new payload
func Make(bts []byte) Payload {
	return Payload{bts}
}

// Push the next message in to the payload
func (p *Payload) Push(bts []byte) {
	p.buf = append(p.buf, bts...)
}

// Pop the next message off the payload
func (p *Payload) Pop() (bts []byte, err error) {
	switch len(p.buf) {
	case 1, 2, 3, 4:
		return bts, fmt.Errorf("error expected at least 5 bytes got %v bytes: %v", len(p.buf), p.buf)
	case 0:
		return bts, nil
	}

	length := 1 + binary.BigEndian.Uint32(p.buf[1:5])
	bts, p.buf = p.buf[:length], p.buf[length:]
	fmt.Printf("popped message %q:\n% x\n", bts[0], bts)

	return bts, nil
}

// ToBytes returns the full sequence of messages as bytes
func (p *Payload) ToBytes() []byte {
	return p.buf
}
