# go-redis-session
## Implementation of user session written in [Go](https://golang.org/). Session data is stored inside the [Redis](http://redis.io/) database.

### USAGE
1. Import dependency

```go
import "github.com/go-redis/redis"
```

`$ go get github.com/go-redis/redis`

2. Create Redis client

```go
options := &redis.Options{
	Addr:     "redishost:6379",
	Password: "secret",
	DB:       0,
}

client := redis.NewClient(options)
```

3. Create Session Store

```go
store := NewStore(client)
```

4. Supported operations

```go
session, err := store.Create(sessionID, time.Duration(10)*time.Second)

err = store.Delete(sessionID)

err = store.Save(session)

session, err := store.Find(sessionID)

err = session.Add(key, value)

name := new(string)
err = session.Get(key, name)
```


### TESTING
1. Make sure docker daemon is running.
2. Run `go test ./...`

