package session

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

const (
	validKey = "__valid__"
)

// NotFound is an interface for errors returned when requested resource cannot be found.
type NotFound interface {
	Name() string
	Key() string
}

// NewNotFoundError constructor function for NotFoundError errors.
func NewNotFoundError(id string) *NotFoundError {
	return &NotFoundError{
		name: "session",
		id:   id,
	}
}

// NotFoundError error returned when Session cannot be found.
type NotFoundError struct {
	name string
	id   string
}

// Name returns name of the resource that cannot be found.
func (e *NotFoundError) Name() string {
	return e.name
}

// Key returns key / id of the resource that cannot be found.
func (e *NotFoundError) Key() string {
	return e.id
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%v with id '%v' cannot be found", e.Name(), e.Key())
}

// Session represents user session.
type Session struct {
	ID     string
	values map[string]string
	valid  time.Duration
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

// Get reats value from session.
func (s *Session) Get(key string, value interface{}) error {
	valStr, ok := s.values[key]
	if !ok {
		return fmt.Errorf("not found")
	}

	if err := json.Unmarshal([]byte(valStr), value); err != nil {
		return err
	}

	return nil
}

// Store is a struct that can be used to create, store, search for sessions.
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

	session := &Session{
		ID:     ID,
		values: make(map[string]string, 0),
		valid:  valid,
	}

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
	log.Printf("Create HGetAll %v", values)

	if len(values) == 0 {
		return nil, NewNotFoundError(ID)
	}

	secondsStr := values[validKey]
	seconds, err := strconv.ParseInt(secondsStr, 10, 32)
	if err != nil {
		return nil, err
	}
	log.Printf("Find Duration secs %v", seconds)

	valid := time.Duration(seconds) * time.Second
	log.Printf("Find Duration %v", valid)
	return &Session{
		ID:     ID,
		values: values,
		valid:  valid,
	}, nil
}

// Save persists given session
func (s *Store) Save(session *Session) error {

	seconds := session.valid.Seconds()
	log.Printf("Duration %v", session.valid)

	sessionInit := make(map[string]interface{}, 0)
	for key, value := range session.values {
		sessionInit[key] = value
	}

	if _, err := s.client.HMSet(session.ID, sessionInit).Result(); err != nil {
		return err
	}

	if _, err := s.client.Expire(session.ID, time.Duration(seconds)*time.Second).Result(); err != nil {
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
		return NewNotFoundError(ID)
	}
	return nil
}

// Close closes session store.
func (s *Store) Close() error {
	return s.client.Close()
}
