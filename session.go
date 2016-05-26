package redissession

import (
	//"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"gopkg.in/redis.v3"
	"strconv"
	"time"
)

const (
	valid_for             = "__##__valid_for__##__"
	defaultSessionTimeout = time.Hour
	empty                 = ""
)

type SessionStore struct {
	redisClient *redis.Client
}

func (store *SessionStore) NewSession(valid time.Duration) (*session, error) {
	sessionId, err := generateSessionId(20)
	if err != nil {
		return nil, err
	}

	sess := session{client: store.redisClient, id: sessionId, valid: valid}
	sess.Add(valid_for, strconv.Itoa(int(valid.Seconds())))
	return &sess, nil
}

func (store *SessionStore) FindSession(sessionId string) (*session, error) {
	sess := session{client: store.redisClient, id: sessionId}
	_, err := sess.Get(valid_for)

	if err == nil {
		sess.fillLifeDuration()
	}

	return &sess, err

}

func NewSessionStore() (SessionStore, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := redisClient.Ping().Result()
	fmt.Println("PING: ", pong, err)

	return SessionStore{redisClient: redisClient}, err
}

type session struct {
	client *redis.Client
	id     string
	valid  time.Duration
}

type sessionNotFound struct {
	sessionId string
}

func (err sessionNotFound) Error() string {
	return fmt.Sprintf("Session with id: %s not found", err.sessionId)
}

type operationFailed struct {
	operation string
	cause     error
}

func (err operationFailed) Error() string {
	return fmt.Sprintf("Operation: %s failed because of: %s", err.operation, err.cause)
}

// ------- TYPES -------

func (session *session) fillLifeDuration() {
	secondsStr, err := session.Get(valid_for)

	if err == nil {
		seconds, e := strconv.Atoi(secondsStr)
		if e != nil {
			session.valid = defaultSessionTimeout
		} else {
			session.valid = time.Duration(seconds) * time.Second
		}

	}
}

// ------- UTIL FUNCTIONS -------

func generateSessionId(idLen int) (string, error) {
	b := make([]byte, idLen)
	_, err := rand.Read(b)
	if err != nil {
		return empty, errors.New("Cannont generate session ID")
	}

	return string(b[:]), nil
}

// ------- SESSION FUNCTIONS -------

func (session *session) Add(name, value string) error {
	command := session.client.HMSet(session.id, name, value)
	//fmt.Println("[Session Add (HMSet)]: ", command.String())

	if err := command.Err(); err != nil {
		return operationFailed{operation: "HMSet", cause: err}
	}

	return nil
}

func (session *session) Get(name string) (string, error) {
	command := session.client.HMGet(session.id, name)

	if err := command.Err(); err != nil {
		return empty, operationFailed{operation: "HMGet", cause: err}
	}

	if val := command.Val(); val != nil && len(val) > 0 && val[0] != nil {
		return val[0].(string), nil
	}

	return empty, nil
}