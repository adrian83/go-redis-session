package session

import (
	"testing"
	"time"
)

const (
	// few consts used in tests
	key   = "name"
	value = "John"
)

var (
	config = Config{
		DB:       0,
		Password: "",
		Host:     "localhost",
		Port:     6390,
		IDLength: 50,
	}
)

func TestRedisSessionImplementsSessionInterface(t *testing.T) {
	var _ Session = redisSession{}
}

func TestStoreCreation(t *testing.T) {

	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("Store cannot be created because of: %v", err)
	}

	// cleanup
	if err = store.Close(); err != nil {
		t.Fatalf("Cannot close Store because of: %v", err)
	}
}

func TestSessionCreation(t *testing.T) {

	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("Store cannot be created because of: %v", err)
	}

	_, err = store.NewSession(time.Duration(10) * time.Second)
	if err != nil {
		t.Fatalf("Session cannot be created because of: %v", err)
	}

	// cleanup
	if err = store.Close(); err != nil {
		t.Fatalf("Cannot close Store because of: %v", err)
	}
}

func TestFindNotExistingSession(t *testing.T) {

	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("Store cannot be created because of: %v", err)
	}

	_, err = store.FindSession("abc")
	if err == nil {
		t.Fatalf("For some reason session exists")
	}

	// cleanup
	if err = store.Close(); err != nil {
		t.Fatalf("Cannot close Store because of: %v", err)
	}
}

func TestFindExistingSession(t *testing.T) {

	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("Store cannot be created because of: %v", err)
	}

	session, err := store.NewSession(time.Duration(10) * time.Second)
	if err != nil {
		t.Fatalf("Session cannot be created because of: %v", err)
	}

	session.Add(key, value)

	err = store.SaveSession(session)
	if err != nil {
		t.Fatalf("Session cannot be saved because of: %v", err)
	}

	session2, err := store.FindSession(session.ID())
	if err != nil {
		t.Fatalf("Session cannot be found because of: %v", err)
	}

	name, _ := session2.Get(key)

	if value != name {
		t.Fatalf("Invalid value in session. Should be '%v', but is '%v'", value, name)
	}

	// cleanup
	if err = store.Close(); err != nil {
		t.Fatalf("Cannot close Store because of: %v", err)
	}
}

func TestSessionProlongation(t *testing.T) {

	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("Store cannot be created because of: %v", err)
	}

	session, err := store.NewSession(time.Duration(3) * time.Second)
	if err != nil {
		t.Fatalf("Session cannot be created because of: %v", err)
	}

	time.Sleep(time.Duration(2) * time.Second)

	session.Add(key, value)

	err = store.SaveSession(session)
	if err != nil {
		t.Fatalf("Session cannot be saved because of: %v", err)
	}

	time.Sleep(time.Duration(2) * time.Second)

	_, err = store.FindSession(session.ID())
	if err != nil {
		t.Fatalf("Session cannot be found because of: %v", err)
	}

	// cleanup
	if err = store.Close(); err != nil {
		t.Fatalf("Cannot close Store because of: %v", err)
	}

}
