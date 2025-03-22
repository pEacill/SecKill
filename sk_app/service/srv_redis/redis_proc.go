package srv_redis

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/pEacill/SecKill/sk_app/config"
	"github.com/pEacill/SecKill/sk_app/model"
)

func WriteHandle() {
	for {
		fmt.Println("Write data to redis")
		req := <-config.SkAppContext.SecReqChan
		fmt.Println("Access Time: ", req.AccessTime)
		conn := conf.Redis.RedisConn

		data, err := json.Marshal(req)
		if err != nil {
			log.Printf("JSON Marshal error: %v, from request: %v", err, req)
			continue
		}

		err = conn.LPush(conf.Redis.Proxy2layerQueueName, string(data)).Err()
		if err != nil {
			log.Printf("Lpush request failed, Error: %v, req: %v.", err, req)
			continue
		}

		log.Printf("Lpush req success. req: %v", string(data))
	}
}

func ReadHandle() {
	for {
		conn := conf.Redis.RedisConn

		data, err := conn.BRPop(time.Second, conf.Redis.Layer2proxyQueueName).Result()
		if err != nil {
			log.Printf("Read Redis error: %v", err)
			continue
		}

		var result *model.SecResult
		err = json.Unmarshal([]byte(data[1]), &result)
		if err != nil {
			log.Printf("JSON.Unmarshal failed, error: %v", err)
			continue
		}

		userKey := fmt.Sprintf("%d_%d", result.UserId, result.ProductId)
		fmt.Println("UserKey: ", userKey)
		config.SkAppContext.UserConnMapLock.Lock()
		resultChan, ok := config.SkAppContext.UserConnMap[userKey]
		config.SkAppContext.UserConnMapLock.Unlock()
		if !ok {
			log.Printf("{User %v} not found", userKey)
			continue
		}
		log.Println("Request result send to chan")

		resultChan <- result
		log.Printf("Request result send to chan success, userKey: %v", userKey)
	}
}
