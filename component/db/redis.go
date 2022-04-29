package db

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

func NewRedisClient(addr, password string, dbNum uint8) (db *redis.Client, err error) {
	db = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,   // no password set
		DB:       int(dbNum), // use default DB
	})

	_, err = db.Ping().Result()
	err = errors.WithStack(err)
	if err != nil {
		return
	}
	return
}
