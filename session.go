package session

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	redis "gopkg.in/redis.v5"
)

const (
	alowedIDChars = "abcdefghijklmnopqrstuvwxyz0123456789"

	valid = "__valid__"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// Session interface represents user session.
type Session interface {
	ID() string
	Add(name string, value string)
	Get(name string) (string, bool)
	Values() map[string]string
	Valid() time.Duration
}

// OperationFailed represents error
type OperationFailed struct {
	operation string
	cause     error
}

// Error abc
func (err OperationFailed) Error() string {
	return fmt.Sprintf("Operation: %s failed because of: %s", err.operation, err.cause)
}

// NotFound abc
type NotFound struct {
	id string
}

// Error abc
func (err NotFound) Error() string {
	return fmt.Sprintf("Session with id %s not found", err.id)
}

// Config contains all the values used by sessions store.
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	IDLength int
}

// NewStore creates new Store struct based on the provided configuration.
func NewStore(config Config) (Store, error) {
	options := &redis.Options{
		Addr:     config.Host + ":" + strconv.Itoa(config.Port),
		Password: config.Password,
		DB:       config.DB,
	}

	client := redis.NewClient(options)

	_, err := client.Ping().Result()

	return Store{client: client, config: config}, err
}

// Store is a struct used for creating, updateing and searching for sessions.
type Store struct {
	client *redis.Client
	config Config
}

// NewSession creates new session with given life length.
func (s *Store) NewSession(valid time.Duration) (Session, error) {
	return &redisSession{
		client: s.client,
		id:     randomString(s.config.IDLength),
		valid:  valid,
		values: make(map[string]string)}, nil
}

// Close cleans up Store resources.
func (s *Store) Close() error {
	return s.client.Close()
}

// FindSession function is used to get session by its id.
func (s *Store) FindSession(sessionID string) (Session, error) {

	values, err := s.client.HGetAll(sessionID).Result()
	if err != nil {
		return nil, OperationFailed{operation: "HGetAll", cause: err}
	}

	if len(values) == 0 {
		return nil, NotFound{id: sessionID}
	}

	secondsStr := values[valid]
	seconds, err := strconv.Atoi(secondsStr)
	if err != nil {
		return nil, errors.New("Cannot read session duration")
	}

	session := redisSession{
		id:     sessionID,
		client: s.client,
		valid:  time.Duration(seconds) * time.Second,
		values: values,
	}

	return session, err
}

// SaveSession saves given session into Redis.
func (s *Store) SaveSession(session Session) error {

	seconds := session.Valid().Seconds()
	session.Values()[valid] = strconv.Itoa(int(seconds))

	if _, err := s.client.HMSet(session.ID(), session.Values()).Result(); err != nil {
		return OperationFailed{operation: "HMSet", cause: err}
	}

	if _, err := s.client.Expire(session.ID(), session.Valid()).Result(); err != nil {
		return OperationFailed{operation: "Expire", cause: err}
	}

	return nil
}

type redisSession struct {
	id     string
	client *redis.Client
	valid  time.Duration
	values map[string]string
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
func (s redisSession) Add(name string, value string) {
	s.values[name] = value
}

// Get abc. The func is part of Session interface.
func (s redisSession) Get(name string) (string, bool) {
	val, ok := s.values[name]
	return val, ok
}

// Values abc. The func is part of Session interface.
func (s redisSession) Values() map[string]string {
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
