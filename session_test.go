package redissession

import (
	"testing"
	//"time"
)

func TestRedisSessionImplementsSessionInterface(t *testing.T) {

	var session Session = redisSession{}
	t.Logf("redisSession implements Session %v", session)

}

/*
func TestSessionStoreCreation(t *testing.T) {

	t.Logf("Given the need to create a session store.")
	t.Logf("\tWhen creating new session store")

	sessionStore, err := NewSessionStore()

	if err != nil {
		t.Fatalf("\t\tThen error is returned: %s", err.Error())
	}

	t.Logf("\t\tThen session store is returned: %v", sessionStore)

}


func TestSessionCreation(t *testing.T) {

	t.Logf("Given the need to create a session.")
	t.Logf("\tWhen creating new session")

	sessionId := NewSession(time.Hour)
	_, err := FindSession(sessionId)

	if err != nil {
		t.Fatalf("\t\tThen error is returned: %s", err.Error())
	}

	t.Logf("\t\tThen session is returned with id: %s", sessionId)

}

func TestSessionReadWriteOperations(t *testing.T) {

	key := "name"
	value := "John"

	t.Logf("Given the need to test read/write session operations.")
	sessionId := NewSession(time.Hour)

	t.Logf("\tWhen saving data do session")
	session, _ := FindSession(sessionId)
	err := session.Add(key, value)

	if err != nil {
		t.Fatalf("\t\tThen error is returned: %s", err)
	} else {
		t.Logf("\t\tThen no error is returned")
	}

	t.Logf("\tWhen reading data from session")
	session, _ = FindSession(sessionId)
	name, err := session.Get(key)

	if err != nil {
		t.Fatalf("\t\tThen error is returned: %s", err.Error())
	}

	if name == value {
		t.Logf("\t\tThen data is the same as it should '%s' == '%s'", name, value)
	} else {
		t.Fatalf("\t\tThen data is not the same as it should '%s' == '%s'", name, value)
	}

}

func TestSessionTimeout(t *testing.T) {

	t.Logf("Given the need to test session timeout.")

	t.Logf("\tWhen session with short lifetime is created")
	sessionId := NewSession(1 * time.Second)

	time.Sleep(1 * time.Second)
	_, err := FindSession(sessionId)

	if err != nil {
		t.Fatalf("\t\tThen after some time the session should be dead but it isn't")
	}

	t.Logf("\t\tThen after some time the session is dead")

}

func TestSessionProlongation(t *testing.T) {

	key := "name"
	value := "John"

	t.Logf("Given the need to test session prolongation.")

	t.Logf("\tWhen session with short lifetime is created")

	sessionId := NewSession(2 * time.Second)
	time.Sleep(1 * time.Second)

	t.Logf("\tAnd during it's lifetime it is used")
	session, _ := FindSession(sessionId)
	session.Add(key, value)

	time.Sleep(1 * time.Second)
	session, err := FindSession(sessionId)

	if err != nil {
		t.Fatalf("\t\tThen after some time the session is dead")
	}

	name, err := session.Get(key)

	if value == name {
		t.Logf("\t\tThen after some time the session is still alive and contains valid data '%s' == '%s'", value, name)
	} else {
		t.Fatalf("\t\tThen after some time the session is still alive but it contains invalid data '%s' != '%s'", value, name)
	}
}
*/
