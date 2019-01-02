package session

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

var (
	// ErrSessionNotFound error returned when requested session cannot be found.
	ErrSessionNotFound = fmt.Errorf("session not found")
	// ErrValueNotFound error returned when requested value cannot be found
	ErrValueNotFound = fmt.Errorf("value not found")
)

// Session session is an interface for sessions.
type Session interface {
	ID() string
	Add(key string, value interface{}) error
	Get(key string, value interface{}) error
	Remove(key string)
	Values() map[string]string
	Removed() []string
}

func newSession(ID string) *session {
	return &session{
		id:      ID,
		values:  make(map[string]string, 0),
		removed: make([]string, 0),
	}
}

// session represents user session.
type session struct {
	id      string
	values  map[string]string
	removed []string
}

func (s *session) ID() string {
	return s.id
}

func (s *session) Values() map[string]string {
	return s.values
}

func (s *session) Removed() []string {
	return s.removed
}

func (s *session) toRedisDict() map[string]interface{} {
	sessionInit := make(map[string]interface{}, 0)
	for key, value := range s.values {
		sessionInit[key] = value
	}
	return sessionInit
}

// Add adds value to session.
func (s *session) Add(key string, value interface{}) error {
	bts, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.values[key] = string(bts)
	return nil
}

// Get reads value from session.
func (s *session) Get(key string, value interface{}) error {
	valStr, ok := s.values[key]
	if !ok {
		return ErrValueNotFound
	}

	if err := json.Unmarshal([]byte(valStr), value); err != nil {
		return err
	}

	return nil
}

// Remove removes value with given key from session.
func (s *session) Remove(key string) {
	s.removed = append(s.removed, key)
	delete(s.values, key)
}

// Store is an interface for session stores.
type Store interface {
	Create(ID string) (Session, error)
	Find(ID string) (Session, error)
	Save(session Session) error
	Delete(ID string) error
	Close() error
}

// store is a struct that can be used to create, store and search for sessions.
type store struct {
	client *redis.Client
	valid  time.Duration
}

// NewStore creates new Store struct.
func NewStore(client *redis.Client, validSec int) Store {
	return &store{
		client: client,
		valid:  time.Duration(validSec) * time.Second,
	}
}

// Create returns new Session with given ID or error if something went wrong.
func (s *store) Create(ID string) (Session, error) {

	session := newSession(ID)
	session.Add("exists", true)

	if _, err := s.client.HMSet(ID, session.toRedisDict()).Result(); err != nil {
		return nil, err
	}

	if _, err := s.client.Expire(ID, s.valid).Result(); err != nil {
		return nil, err
	}

	return session, nil
}

// Find returns Session with given ID if it exist. Error otherwise.
func (s *store) Find(ID string) (Session, error) {

	values, err := s.client.HGetAll(ID).Result()
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, ErrSessionNotFound
	}

	session := newSession(ID)
	session.values = values

	return session, nil
}

// Save persists given session
func (s *store) Save(session Session) error {

	removed := session.Removed()
	if len(removed) > 0 {
		if _, err := s.client.HDel(session.ID(), removed...).Result(); err != nil {
			return err
		}
	}

	sessionVals := make(map[string]interface{}, 0)
	for key, value := range session.Values() {
		sessionVals[key] = value
	}

	if _, err := s.client.HMSet(session.ID(), sessionVals).Result(); err != nil {
		return err
	}

	if _, err := s.client.Expire(session.ID(), s.valid).Result(); err != nil {
		return err
	}

	return nil
}

// Delete removes session with given ID from store.
func (s *store) Delete(ID string) error {

	count, err := s.client.Del(ID).Result()
	if err != nil {
		return err
	}

	if count < 1 {
		return ErrSessionNotFound
	}

	return nil
}

// Close closes session store.
func (s *store) Close() error {
	return s.client.Close()
}
