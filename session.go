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

// Store interface represents session store.
type Store interface {
	NewSession(valid time.Duration) (Session, error)
	Close() error
	FindSession(sessionID string) (Session, error)
	SaveSession(session Session) error
}

// OperationFailed represents error occured during execution of redis commands.
type OperationFailed struct {
	operation string
	cause     error
}

// Error returns string representation of OperationFailed error struct.
func (err OperationFailed) Error() string {
	return fmt.Sprintf("Operation: %s failed because of: %s", err.operation, err.cause)
}

// NotFound error returned when session cannot be found.
type NotFound struct {
	id string
}

// Error returns string representation of NotFound error struct.
func (err NotFound) Error() string {
	return fmt.Sprintf("Session with id %s not found", err.id)
}

// Config contains all the values needed to create sessions store.
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	IDLength int
}

// NewStore creates new session store struct based on the provided configuration.
func NewStore(config Config) (Store, error) {
	options := &redis.Options{
		Addr:     config.Host + ":" + strconv.Itoa(config.Port),
		Password: config.Password,
		DB:       config.DB,
	}

	client := redis.NewClient(options)

	_, err := client.Ping().Result()

	return &redisStore{client: client, config: config}, err
}

// Store is a struct used for creating, updateing and searching for sessions.
type redisStore struct {
	client *redis.Client
	config Config
}

// NewSession creates new session with given life length.
func (s *redisStore) NewSession(valid time.Duration) (Session, error) {
	return &redisSession{
		client: s.client,
		id:     randomString(s.config.IDLength),
		valid:  valid,
		values: make(map[string]string)}, nil
}

// Close closes the connection with Redis.
func (s *redisStore) Close() error {
	return s.client.Close()
}

// FindSession function is used to get session by its id.
func (s *redisStore) FindSession(sessionID string) (Session, error) {

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
func (s *redisStore) SaveSession(session Session) error {

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

// redisSession is an implementation of a Session interface in which session is stored in Redis.
type redisSession struct {
	id     string
	client *redis.Client
	valid  time.Duration
	values map[string]string
}

// String returns string representation of the redisSession.
func (s *redisSession) String() string {
	return fmt.Sprintf("redisSession { id: %v, valid: %v, values: %v }", s.id, s.valid, s.values)
}

// ID returns id of the session.
func (s redisSession) ID() string {
	return s.id
}

// Add adds key-value pair to the session.
func (s redisSession) Add(name string, value string) {
	s.values[name] = value
}

// Get returns the value stored in session under given key.
func (s redisSession) Get(name string) (string, bool) {
	val, ok := s.values[name]
	return val, ok
}

// Values returns map with values stored in session.
func (s redisSession) Values() map[string]string {
	return s.values
}

// Valid returns duration for how long session is valid.
func (s redisSession) Valid() time.Duration {
	return s.valid
}

// randomString returns random string with given length.
func randomString(strLen int) string {
	result := make([]byte, strLen)
	l := len(alowedIDChars)
	for i := 0; i < strLen; i++ {
		result[i] = alowedIDChars[rand.Intn(l)]
	}
	return string(result)
}
