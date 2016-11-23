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

func TestSessionStoreCreation(t *testing.T) {

	sessionStore, err := NewSessionStore(config)
	if err != nil {
		t.Fatalf("SessionStore cannot be created because of: %v", err)
	}

	// cleanup
	if err = sessionStore.Close(); err != nil {
		t.Fatalf("Cannot close SessionStore because of: %v", err)
	}
}

func TestSessionCreation(t *testing.T) {

	sessionStore, err := NewSessionStore(config)
	if err != nil {
		t.Fatalf("SessionStore cannot be created because of: %v", err)
	}

	_, err = sessionStore.NewSession(time.Duration(10) * time.Second)
	if err != nil {
		t.Fatalf("Session cannot be created because of: %v", err)
	}

	// cleanup
	if err = sessionStore.Close(); err != nil {
		t.Fatalf("Cannot close SessionStore because of: %v", err)
	}
}

func TestFindNotExistingSession(t *testing.T) {

	sessionStore, err := NewSessionStore(config)
	if err != nil {
		t.Fatalf("SessionStore cannot be created because of: %v", err)
	}

	_, err = sessionStore.FindSession("abc")
	if err == nil {
		t.Fatalf("For some reason session exists")
	}

	// cleanup
	if err = sessionStore.Close(); err != nil {
		t.Fatalf("Cannot close SessionStore because of: %v", err)
	}
}

func TestFindExistingSession(t *testing.T) {

	sessionStore, err := NewSessionStore(config)
	if err != nil {
		t.Fatalf("SessionStore cannot be created because of: %v", err)
	}

	session, err := sessionStore.NewSession(time.Duration(10) * time.Second)
	if err != nil {
		t.Fatalf("Session cannot be created because of: %v", err)
	}

	session.Add(key, value)

	err = sessionStore.SaveSession(session)
	if err != nil {
		t.Fatalf("Session cannot be saved because of: %v", err)
	}

	session2, err := sessionStore.FindSession(session.ID())
	if err != nil {
		t.Fatalf("Session cannot be found because of: %v", err)
	}

	name, _ := session2.Get(key)

	if value != name {
		t.Fatalf("Invalid value in session. Should be '%v', but is '%v'", value, name)
	}

	// cleanup
	if err = sessionStore.Close(); err != nil {
		t.Fatalf("Cannot close SessionStore because of: %v", err)
	}
}

func TestSessionProlongation(t *testing.T) {

	sessionStore, err := NewSessionStore(config)
	if err != nil {
		t.Fatalf("SessionStore cannot be created because of: %v", err)
	}

	session, err := sessionStore.NewSession(time.Duration(3) * time.Second)
	if err != nil {
		t.Fatalf("Session cannot be created because of: %v", err)
	}

	time.Sleep(time.Duration(2) * time.Second)

	session.Add(key, value)

	err = sessionStore.SaveSession(session)
	if err != nil {
		t.Fatalf("Session cannot be saved because of: %v", err)
	}

	time.Sleep(time.Duration(2) * time.Second)

	_, err = sessionStore.FindSession(session.ID())
	if err != nil {
		t.Fatalf("Session cannot be found because of: %v", err)
	}

	// cleanup
	if err = sessionStore.Close(); err != nil {
		t.Fatalf("Cannot close SessionStore because of: %v", err)
	}

}
