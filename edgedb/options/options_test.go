package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHost(t *testing.T) {
	opts := FromDSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, "localhost", opts.Host)
}

func TestParsePort(t *testing.T) {
	opts := FromDSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, 5656, opts.Port)
}

func TestParseUser(t *testing.T) {
	opts := FromDSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, "me", opts.User)
}

func TestParseDatabase(t *testing.T) {
	opts := FromDSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, "somedb", opts.Database)
}

func TestParsePassword(t *testing.T) {
	opts := FromDSN("edgedb://me:secret@localhost:5656/somedb")
	assert.Equal(t, "secret", opts.Password)

}
