package api

import (
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
)

func TestModuleBlacklist(t *testing.T) {
	c, err := NewClient(true, "test.db")
	assert.NoError(t, err)

	b, err := c.IsTokenBlacklisted([]byte("test"))
	assert.NoError(t, err)
	assert.False(t, b)

	assert.NoError(t, c.BlacklistToken([]byte("test")))

	b, err = c.IsTokenBlacklisted([]byte("test"))
	assert.NoError(t, err)
	assert.True(t, b)

	assert.NoError(t, c.Close())

	os.Remove("test.db")
}

func TestModuleLastSeen(t *testing.T) {
	c, err := NewClient(true, "test.db")
	assert.NoError(t, err)

	userId := uuid.Must(uuid.NewV4())
	tt, err := c.GetUserLastSeenTime(userId)
	assert.NoError(t, err)
	assert.True(t, time.Now().After(tt))

	assert.NoError(t, c.SaveUserLastSeenTime(userId))

	tt, err = c.GetUserLastSeenTime(userId)
	assert.NoError(t, err)
	assert.True(t, time.Now().Add(-1*time.Minute).Before(tt))

	assert.NoError(t, c.Close())

	os.Remove("test.db")
}
