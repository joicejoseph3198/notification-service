package config

import (
	"context"

	"github.com/redis/go-redis/v9"
)



var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis(){
	RedisClient = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
}

