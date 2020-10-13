package edgedb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	opts := DSN("edgedb://user_with_password:secret@localhost/edgedb")
	conn, err := Connect(opts)
	assert.Nil(t, err)

	result, err := conn.QueryOneJSON("SELECT 'It worked!';")
	assert.Nil(t, err)
	assert.Equal(t, `"It worked!"`, string(result))

}
