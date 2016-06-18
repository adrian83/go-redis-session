package redissession

import (
	"encoding/json"
	"fmt"
	"gopkg.in/redis.v3"
	"math/rand"
	"strconv"
	"time"
)

const (
	alowedIDChars = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

/* Interfaces */

// Session interface represents user session.
type Session interface {
	ID() string
	Add(name string, value interface{})
	Get(name string) (interface{}, bool)
	Values() map[string]interface{}
	Valid() time.Duration
}

/* Structs */

/* Errors */

// OperationFailed abc
type OperationFailed struct {
	operation string
	cause     error
}

// Error abc
func (err OperationFailed) Error() string {
	return fmt.Sprintf("Operation: %s failed because of: %s", err.operation, err.cause)
}

// Config contains all the values used by sessions store.
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int64
	IDLength int
}

// NewSessionStore creates new SessionStore struct based on the provided configuration.
func NewSessionStore(config Config) (SessionStore, error) {
	options := &redis.Options{
		Addr:     config.Host + ":" + strconv.Itoa(config.Port),
		Password: config.Password,
		DB:       config.DB,
	}

	redisClient := redis.NewClient(options)

	_, err := redisClient.Ping().Result()

	return SessionStore{redisClient: redisClient, config: config}, err
}

// SessionStore is a struct used for creating, updateing and searching for sessions.
type SessionStore struct {
	redisClient *redis.Client
	config      Config
}

// NewSession creates new session with given life length.
func (s *SessionStore) NewSession(valid time.Duration) (Session, error) {
	sessionID := randomString(s.config.IDLength)
	sess := redisSession{
		client: s.redisClient,
		id:     sessionID,
		valid:  valid,
		values: make(map[string]interface{})}

	return &sess, nil
}

// Close cleans up SessionStore resources.
func (s *SessionStore) Close() error {
	return s.redisClient.Close()
}

// FindSession function is used to get session by its id.
func (s *SessionStore) FindSession(sessionID string) (Session, error) {

	sessionJSON, err := s.redisClient.Get(sessionID).Result()
	if err != nil {
		return nil, OperationFailed{operation: "Get", cause: err}
	}

	redisSession := new(redisSession)
	if err = json.Unmarshal([]byte(sessionJSON), redisSession); err != nil {
		return nil, err
	}

	return redisSession, err
}

// SaveSession saves given session into Redis.
func (s *SessionStore) SaveSession(session Session) error {

	redisSession := &redisSession{id: session.ID(), values: session.Values(), valid: session.Valid()}

	sessionJSON, err := json.Marshal(redisSession)
	if err != nil {
		return err
	}

	statCmd := s.redisClient.Set(session.ID(), sessionJSON, session.Valid())
	_, err = statCmd.Result()

	if err != nil {
		return OperationFailed{operation: "Set", cause: err}
	}

	return nil
}

type redisSession struct {
	id     string
	client *redis.Client
	valid  time.Duration
	values map[string]interface{}
}

// UnmarshalJSON func is used for unmarshaling redisSession struct.
func (s *redisSession) UnmarshalJSON(b []byte) error {
	f := new(struct {
		ID     string                 `json:"id"`
		Valid  time.Duration          `json:"valid"`
		Values map[string]interface{} `json:"values"`
	})

	if err := json.Unmarshal(b, f); err != nil {
		return err
	}

	s.id = f.ID
	s.valid = f.Valid
	s.values = f.Values
	return nil
}

// MarshalJSON func is used for marshaling redisSession struct.
func (s *redisSession) MarshalJSON() ([]byte, error) {

	bytes, err := json.Marshal(struct {
		ID     string                 `json:"id"`
		Valid  time.Duration          `json:"valid"`
		Values map[string]interface{} `json:"values"`
	}{
		ID:     s.id,
		Valid:  s.valid,
		Values: s.values,
	})

	return bytes, err
}

// String returns representation of the redisSession as a string.
func (s *redisSession) String() string {
	return fmt.Sprintf("redisSession { id: %v, valid: %v, values: %v }", s.id, s.valid, s.values)
}

// Id returns id of the session. The func is part of Session interface.
func (s redisSession) ID() string {
	return s.id
}

// Add adds key-value pair to the session. The func is part of Session interface.
func (s redisSession) Add(name string, value interface{}) {
	s.values[name] = value
}

// Get abc. The func is part of Session interface.
func (s redisSession) Get(name string) (interface{}, bool) {
	val, ok := s.values[name]
	return val, ok
}

// Values abc. The func is part of Session interface.
func (s redisSession) Values() map[string]interface{} {
	return s.values
}

// Valid abc. The func is part of Session interface.
func (s redisSession) Valid() time.Duration {
	return s.valid
}

func randomString(strLen int) string {
	result := make([]byte, strLen)
	l := len(alowedIDChars)
	for i := 0; i < strLen; i++ {
		result[i] = alowedIDChars[rand.Intn(l)]
	}
	return string(result)
}
