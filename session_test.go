package session

import (
	"fmt"
	"testing"
	"time"

	"log"
	"os"

	"github.com/go-redis/redis"
	"github.com/ory/dockertest"
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
	store := NewStore(client)

	if _, err := store.Create("abc", time.Duration(10)*time.Second); err != nil {
		t.Errorf("Session cannot be created because of: %v", err)
	}

}

func TestFindNotExistingSession(t *testing.T) {
	sessionID := "xyz"

	store := NewStore(client)

	if _, err := store.Find(sessionID); err == nil {
		t.Errorf("For some reason session exists")
	}

}

func TestFindExistingSession(t *testing.T) {
	sessionID := "abc"

	store := NewStore(client)

	session, err := store.Create(sessionID, time.Duration(10)*time.Second)
	if err != nil {
		t.Errorf("Session cannot be created because of: %v", err)
	}

	if err = session.Add(key, value); err != nil {
		t.Errorf("Unexpected error while adding value to session")
	}

	if err = store.Save(session); err != nil {
		t.Errorf("Session cannot be saved because of: %v", err)
	}

	session2, err := store.Find(sessionID)
	if err != nil {
		t.Errorf("Session cannot be found because of: %v", err)
	}

	name := new(string)
	if err = session2.Get(key, name); err != nil {
		t.Errorf("Unexpected error while reading value from session")
	}

	if value != *name {
		t.Errorf("Invalid value in session. Should be '%v', but is '%v'", value, *name)
	}
}

func TestDeleteElementFromSession(t *testing.T) {
	sessionID := "bcd"

	store := NewStore(client)

	session, err := store.Create(sessionID, time.Duration(10)*time.Second)
	if err != nil {
		t.Errorf("Session cannot be created because of: %v", err)
	}

	if err = session.Add(key, value); err != nil {
		t.Error("Unexpected error while adding value to session")
	}

	if err = store.Save(session); err != nil {
		t.Errorf("Session cannot be saved because of: %v", err)
	}

	session.Remove(key)

	if err = store.Save(session); err != nil {
		t.Errorf("Session cannot be saved because of: %v", err)
	}

	session2, err := store.Find(sessionID)
	if err != nil {
		t.Errorf("Session cannot be found because of: %v", err)
	}

	name := new(string)
	if err = session2.Get(key, name); err == nil {
		t.Error("Expected error while reading value from session but none was returned")
	}
}

func TestSessionProlongation(t *testing.T) {
	sessionID := "def"

	store := NewStore(client)

	session, err := store.Create(sessionID, time.Duration(3)*time.Second)
	if err != nil {
		t.Errorf("Session cannot be created because of: %v", err)
	}

	time.Sleep(time.Duration(2) * time.Second)

	if err = session.Add(key, value); err != nil {
		t.Errorf("Unexpected error while adding value to session")
	}

	if err = store.Save(session); err != nil {
		t.Errorf("Session cannot be saved because of: %v", err)
	}

	time.Sleep(time.Duration(2) * time.Second)

	_, err = store.Find(sessionID)
	if err != nil {
		t.Errorf("Session cannot be found because of: %v", err)
	}
}

func TestSessionAutoRemoveFunctionality(t *testing.T) {
	sessionID := "klm"

	store := NewStore(client)

	_, err := store.Create(sessionID, time.Duration(1)*time.Second)
	if err != nil {
		t.Errorf("Session cannot be created because of: %v", err)
	}

	_, err = store.Find(sessionID)
	if err != nil {
		t.Errorf("Session cannot be found because of: %v", err)
	}

	time.Sleep(time.Duration(2) * time.Second)

	_, err = store.Find(sessionID)
	if err == nil {
		t.Errorf("Session should not exist")
	}
}

func TestDeleteSession(t *testing.T) {
	sessionID := "def"

	store := NewStore(client)

	session, err := store.Create(sessionID, time.Duration(3)*time.Second)
	if err != nil {
		t.Errorf("Session cannot be created because of: %v", err)
	}

	if err = session.Add("name", "John"); err != nil {
		t.Errorf("Unexpected error while adding value to session")
	}

	if err = store.Save(session); err != nil {
		t.Errorf("Session cannot be saved because of: %v", err)
	}

	err = store.Delete(sessionID)
	if err != nil {
		t.Errorf("Cannot delete session because of: %v", err)
	}

	_, err = store.Find(sessionID)
	if err == nil {
		t.Errorf("Error was expected")
	}
}
