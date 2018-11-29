package session

import (
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

const (
	// connection properties
	db       = 0
	password = ""
	host     = "localhost"
	port     = 6390

	// few consts used in tests
	key   = "name"
	value = "John"
)

var (
	options = &redis.Options{
		Addr:     host + ":" + strconv.Itoa(port),
		Password: password,
		DB:       db,
	}
)

func TestStoreCreation(t *testing.T) {
	client := redis.NewClient(options)

	store := NewStore(client)

	// cleanup
	if err := store.Close(); err != nil {
		t.Errorf("Cannot close Store because of: %v", err)
	}
}

func TestSessionCreation(t *testing.T) {
	client := redis.NewClient(options)

	store := NewStore(client)

	if _, err := store.Create("abc", time.Duration(10)*time.Second); err != nil {
		t.Errorf("Session cannot be created because of: %v", err)
	}

	// cleanup
	if err := store.Close(); err != nil {
		t.Errorf("Cannot close Store because of: %v", err)
	}
}

func TestFindNotExistingSession(t *testing.T) {
	client := redis.NewClient(options)
	sessionID := "xyz"

	store := NewStore(client)

	if _, err := store.Find(sessionID); err == nil {
		t.Errorf("For some reason session exists")
	}

	// cleanup
	if err := store.Close(); err != nil {
		t.Errorf("Cannot close Store because of: %v", err)
	}
}

func TestFindExistingSession(t *testing.T) {
	client := redis.NewClient(options)
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

	// cleanup
	if err = store.Close(); err != nil {
		t.Errorf("Cannot close Store because of: %v", err)
	}
}

func TestSessionProlongation(t *testing.T) {
	client := redis.NewClient(options)
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

	// cleanup
	if err = store.Close(); err != nil {
		t.Errorf("Cannot close Store because of: %v", err)
	}
}

func TestSessionAutoRemoveFunctionality(t *testing.T) {
	client := redis.NewClient(options)
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

	// cleanup
	if err = store.Close(); err != nil {
		t.Errorf("Cannot close Store because of: %v", err)
	}
}

func TestDeleteSession(t *testing.T) {
	client := redis.NewClient(options)
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

	// cleanup
	if err = store.Close(); err != nil {
		t.Errorf("Cannot close Store because of: %v", err)
	}

}
