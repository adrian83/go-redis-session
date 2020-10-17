package session_test

import (
	"testing"
	"time"

	session "github.com/adrian83/go-redis-session"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func TestCreateSession(t *testing.T) {
	// given
	dbClient := dbClientMock{
		hmSetStatus: redis.NewStatusCmd("ok"),
		expireCmd:   redis.NewBoolCmd("ok"),
	}

	sessStore := session.NewStore(&dbClient, 10)

	// when
	sess, err := sessStore.Create("abc")

	// then
	assert.NoError(t, err)
	assert.NotNil(t, sess)
}

type dbClientMock struct {
	hmSetStatus *redis.StatusCmd
	expireCmd   *redis.BoolCmd
	hGetAllCmd  *redis.StringStringMapCmd
	hDelCmd     *redis.IntCmd
	delCmd      *redis.IntCmd
	closeErr    error
}

func (m *dbClientMock) HMSet(key string, fields map[string]interface{}) *redis.StatusCmd {
	return m.hmSetStatus
}

func (m *dbClientMock) Expire(key string, expiration time.Duration) *redis.BoolCmd {
	return m.expireCmd
}

func (m *dbClientMock) HGetAll(key string) *redis.StringStringMapCmd {
	return m.hGetAllCmd
}

func (m *dbClientMock) HDel(key string, fields ...string) *redis.IntCmd {
	return m.hDelCmd
}

func (m *dbClientMock) Del(keys ...string) *redis.IntCmd {
	return m.delCmd
}

func (m *dbClientMock) Close() error {
	return m.closeErr
}
