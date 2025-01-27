package config

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)



var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis(){
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	RedisClient = redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%s", redisHost, redisPort)})
}

