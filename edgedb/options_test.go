package edgedb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHost(t *testing.T) {
	opts := DSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, "localhost", opts.Host)
}

func TestParsePort(t *testing.T) {
	opts := DSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, 5656, opts.Port)
}

func TestParseUser(t *testing.T) {
	opts := DSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, "me", opts.User)
}

func TestParseDatabase(t *testing.T) {
	opts := DSN("edgedb://me@localhost:5656/somedb")
	assert.Equal(t, "somedb", opts.Database)
}

func TestParsePassword(t *testing.T) {
	opts := DSN("edgedb://me:secret@localhost:5656/somedb")
	assert.Equal(t, "secret", opts.Password)
}

func TestDialHost(t *testing.T) {
	opts := Options{Host: "some.com", Port: 1234}
	assert.Equal(t, "some.com:1234", opts.dialHost())

	opts = Options{Port: 1234}
	assert.Equal(t, "localhost:1234", opts.dialHost())

	opts = Options{Host: "some.com"}
	assert.Equal(t, "some.com:5656", opts.dialHost())

	opts = Options{}
	assert.Equal(t, "localhost:5656", opts.dialHost())
}
