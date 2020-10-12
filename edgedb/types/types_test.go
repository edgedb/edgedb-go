package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUIDString(t *testing.T) {
	uuid := UUID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	assert.Equal(t, "00010203-0405-0607-0809-0a0b0c0d0e0f", uuid.String())
}
