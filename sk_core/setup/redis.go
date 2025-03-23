package setup

import (
	"log"

	"github.com/go-redis/redis"
	conf "github.com/pEacill/SecKill/pkg/config"
)

func InitRedis() {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Host + ":6379",
		Password: conf.Redis.Password,
		DB:       conf.Redis.Db,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Connect Redis failed, error: %v", err)
	}

	conf.Redis.RedisConn = client
}
