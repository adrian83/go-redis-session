package session

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis"
)

var (
	// ErrSessionNotFound error returned when requested session cannot be found.
	ErrSessionNotFound = errors.New("session not found")
	// ErrValueNotFound error returned when requested value cannot be found
	ErrValueNotFound = errors.New("value not found")
)

func newSession(ID string) *Session {
	return &Session{
		id:      ID,
		values:  make(map[string]string, 0),
		removed: make([]string, 0),
	}
}

// Session represents user session.
type Session struct {
	id      string
	values  map[string]string
	removed []string
}

func (s *Session) ID() string {
	return s.id
}

func (s *Session) Values() map[string]string {
	return s.values
}

func (s *Session) Removed() []string {
	return s.removed
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
	s.removed = append(s.removed, key)
	delete(s.values, key)
}

type dbClient interface {
	HMSet(key string, fields map[string]interface{}) *redis.StatusCmd
	Expire(key string, expiration time.Duration) *redis.BoolCmd
	HGetAll(key string) *redis.StringStringMapCmd
	HDel(key string, fields ...string) *redis.IntCmd
	Del(keys ...string) *redis.IntCmd
	Close() error
}

// Store is a struct that can be used to create, store and search for sessions.
type Store struct {
	client dbClient
	valid  time.Duration
}

// NewStore creates new Store struct.
func NewStore(client dbClient, validSec int) *Store {
	return &Store{
		client: client,
		valid:  time.Duration(validSec) * time.Second,
	}
}

// Create returns new Session with given ID or error if something went wrong.
func (s *Store) Create(ID string) (*Session, error) {
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
func (s *Store) Find(ID string) (*Session, error) {
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
func (s *Store) Save(session *Session) error {
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
