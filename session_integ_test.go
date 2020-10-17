// +build integration

package session_test

import (
	"fmt"
	"testing"
	"time"

	"log"
	"os"

	session "github.com/adrian83/go-redis-session"
	"github.com/go-redis/redis"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
)

const (
	// docker
	image   = "redis"
	version = "latest"

	// connection properties
	db       = 0
	password = ""
	host     = "localhost"
	port     = "6379/tcp"

	// few consts used in tests
	key   = "name"
	value = "John"
)

var (
	client *redis.Client
)

func uniqueName(prefix string) string {
	return fmt.Sprintf("%v-%v", prefix, time.Now().UnixNano())
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run(image, version, nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, resource.GetPort(port)),
		Password: password,
		DB:       db,
	}

	retryFunc := func() error {
		client = redis.NewClient(options)
		return client.Ping().Err()
	}

	if err = pool.Retry(retryFunc); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestSessionCreation(t *testing.T) {
	// given
	sessionID := uniqueName("abc")
	store := session.NewStore(client, 10)

	// when
	sess, err := store.Create(sessionID)

	// then
	assert.NoError(t, err, "cannot create session")
	assert.NotNil(t, sess, "session not created")
}

func TestFindNotExistingSession(t *testing.T) {
	// given
	sessionID := uniqueName("xyz")
	store := session.NewStore(client, 10)

	// when
	sess, err := store.Find(sessionID)

	// then
	assert.Error(t, err)
	assert.Equal(t, session.ErrSessionNotFound, err)
	assert.Nil(t, sess)
}

func TestFindExistingSession(t *testing.T) {
	// given
	sessionID := uniqueName("def")
	store := session.NewStore(client, 10)

	// when
	session1, err := store.Create(sessionID)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, session1)

	// when
	err = session1.Add(key, value)

	// then
	assert.NoError(t, err)

	err = store.Save(session1)
	assert.NoError(t, err)

	session2, err := store.Find(sessionID)
	assert.NoError(t, err)

	var name string
	err = session2.Get(key, &name)
	assert.NoError(t, err)
	assert.Equal(t, value, name)
}

func TestDeleteElementFromSession(t *testing.T) {
	// given
	sessionID := uniqueName("bcd")
	store := session.NewStore(client, 10)

	// when
	sess, err := store.Create(sessionID)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, sess)

	// when
	err = sess.Add(key, value)

	// then
	assert.NoError(t, err)

	// when
	err = store.Save(sess)

	// then
	assert.NoError(t, err)

	// when
	sess.Remove(key)

	err = store.Save(sess)

	// then
	assert.NotNil(t, sess)

	// when
	sess2, err := store.Find(sessionID)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, sess2)

	// when
	var name string
	err = sess2.Get(key, &name)

	// then
	assert.Error(t, err)
	assert.Equal(t, session.ErrValueNotFound, err)
	assert.Equal(t, "", name)
}

func TestSessionProlongation(t *testing.T) {
	sessionID := uniqueName("fgh")
	store := session.NewStore(client, 3)

	// when
	sess, err := store.Create(sessionID)

	// then
	assert.NoError(t, err)

	// when
	time.Sleep(time.Duration(2) * time.Second)

	err = sess.Add(key, value)

	// then
	assert.NoError(t, err)

	err = store.Save(sess)

	// then
	assert.NoError(t, err)

	time.Sleep(time.Duration(2) * time.Second)

	sess2, err := store.Find(sessionID)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, sess2)
}

func TestSessionAutoRemoveFunctionality(t *testing.T) {
	// given
	sessionID := uniqueName("klm")
	store := session.NewStore(client, 1)

	// when
	sess1, err := store.Create(sessionID)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, sess1)

	sess2, err := store.Find(sessionID)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, sess2)

	time.Sleep(time.Duration(2) * time.Second)

	// when
	sess3, err := store.Find(sessionID)

	// then
	assert.Error(t, err)
	assert.Nil(t, sess3)
}

func TestDeleteSession(t *testing.T) {
	// given
	sessionID := uniqueName("jik")
	store := session.NewStore(client, 3)

	// when
	sess1, err := store.Create(sessionID)

	// then
	assert.NoError(t, err)
	assert.NotNil(t, sess1)

	// when
	err = sess1.Add(key, value)

	// then
	assert.NoError(t, err)

	// when
	err = store.Save(sess1)

	// then
	assert.NoError(t, err)

	// when
	err = store.Delete(sessionID)

	// then
	assert.NoError(t, err)

	// when
	sess2, err := store.Find(sessionID)

	// then
	assert.Error(t, err)
	assert.Nil(t, sess2)
}
