# go-redis-session
## Implementation of user session written in [Go](https://golang.org/). Session data is stored inside the [Redis](http://redis.io/) database.

### USAGE
1. Import dependency

import "github.com/go-redis/redis"

go get github.com/go-redis/redis

2. Crete Redis client

	options := &redis.Options{
		Addr:     "redishost:6379",
		Password: "secret",
		DB:       0,
	}

    client := redis.NewClient(options)

3. Creating Session Store

store := NewStore(client)

4. Supported operations

	session, err := store.Create(sessionID, time.Duration(10)*time.Second)

    err = store.Delete(sessionID)

    err = store.Save(session)

    session, err := store.Find(sessionID)


	err = session.Add(key, value)

    name := new(string)
    err = session.Get(key, name)





### TESTING
1. Executing tests require running docker daemon.
2. Run go test ./...

