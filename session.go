package session

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const (
	validKey = "__valid__"
)

var (
	// ErrSessionNotFound error returned when requested session cannot be found.
	ErrSessionNotFound = fmt.Errorf("session not found")
	// ErrValueNotFound error returned when requested value cannot be found
	ErrValueNotFound = fmt.Errorf("value not found")
)

func newSession(ID string, valid time.Duration) *Session {
	return &Session{
		ID:      ID,
		values:  make(map[string]string, 0),
		removed: make([]string, 0),
		valid:   valid,
	}
}

// Session represents user session.
type Session struct {
	ID      string
	values  map[string]string
	removed []string
	valid   time.Duration
}

func (s *Session) toRedisDict() map[string]interface{} {
	sessionInit := make(map[string]interface{}, 0)
	for key, value := range s.values {
		sessionInit[key] = value
	}
	return sessionInit
}

// Add adds value to session.
func (s *Session) Add(key string, value interface{}) error {
	bts, err := json.Marshal(value)
	if err != nil {
		return err
	}
	s.values[key] = string(bts)
	return nil
}

// Get reads value from session.
func (s *Session) Get(key string, value interface{}) error {
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
func (s *Session) Remove(key string) {
	if key != validKey {
		s.removed = append(s.removed, key)
		delete(s.values, key)
	}
}

// Store is a struct that can be used to create, store and search for sessions.
type Store struct {
	client *redis.Client
}

// NewStore creates new Store struct.
func NewStore(client *redis.Client) *Store {
	return &Store{client: client}
}

// Create returns new Session with given ID that will be persisted for
// given duration or error if something went wrong.
func (s *Store) Create(ID string, valid time.Duration) (*Session, error) {

	session := newSession(ID, valid)
	session.Add(validKey, valid.Seconds())

	_, err := s.client.HMSet(ID, session.toRedisDict()).Result()
	if err != nil {
		return nil, err
	}

	if _, err := s.client.Expire(ID, valid).Result(); err != nil {
		return nil, err
	}

	return session, nil
}

// Find returns Session with given ID if it exist. Error otherwise.
func (s *Store) Find(ID string) (*Session, error) {

	values, err := s.client.HGetAll(ID).Result()
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, ErrSessionNotFound
	}

	secondsStr := values[validKey]
	seconds, err := strconv.ParseInt(secondsStr, 10, 32)
	if err != nil {
		return nil, err
	}

	valid := time.Duration(seconds) * time.Second

	session := newSession(ID, valid)
	session.values = values

	return session, nil
}

// Save persists given session
func (s *Store) Save(session *Session) error {

	if len(session.removed) > 0 {
		if _, err := s.client.HDel(session.ID, session.removed...).Result(); err != nil {
			return err
		}
	}

	if _, err := s.client.HMSet(session.ID, session.toRedisDict()).Result(); err != nil {
		return err
	}

	if _, err := s.client.Expire(session.ID, session.valid).Result(); err != nil {
		return err
	}

	return nil
}

// Delete removes session with given ID from store.
func (s *Store) Delete(ID string) error {

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
func (s *Store) Close() error {
	return s.client.Close()
}
