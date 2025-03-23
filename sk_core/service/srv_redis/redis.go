package srv_redis

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/pEacill/SecKill/sk_core/config"
	"github.com/pEacill/SecKill/sk_core/model"
)

func RunProcess() {
	for i := 0; i < conf.SecKill.CoreReadRedisGoroutineNum; i++ {
		go HandleReader()
	}

	for i := 0; i < conf.SecKill.CoreWriteRedisGoroutineNum; i++ {
		go HandleWrite()
	}

	for i := 0; i < conf.SecKill.CoreHandleGoroutineNum; i++ {
		go HandleUser()
	}

	log.Println("All process goroutine started.")
	return
}

func HandleReader() {
	log.Printf("Read goroutine running, read queue: %v", conf.Redis.Proxy2layerQueueName)

	for {
		conn := conf.Redis.RedisConn
		for {
			data, err := conn.BRPop(time.Second, conf.Redis.Proxy2layerQueueName).Result()
			if err != nil {
				continue
			}
			log.Printf("BRPop from proxy to layer queue, data: %v", data)

			var req model.SecRequest
			err = json.Unmarshal([]byte(data[1]), &req)
			if err != nil {
				log.Printf("Unmarshal to SecrKill request failed, with error: %v", err)
				continue
			}

			nowTime := time.Now().Unix()
			fmt.Println(nowTime, " ", req.SecTime, " ", 100)
			if nowTime-req.SecTime >= int64(conf.SecKill.MaxRequestWaitTimeout) {
				log.Printf("{Req: %v} is expire.", req)
				continue
			}

			timer := time.NewTicker(time.Millisecond * time.Duration(conf.SecKill.CoreWaitResultTimeout))

			select {
			case config.SecLayerCtx.Read2HandleChan <- &req:
			case <-timer.C:
				log.Printf("{Req: %v} send to handle chan timeout.", req)
				break
			}
		}
	}
}

func HandleWrite() {
	log.Println("Write goroutine running.")

	for res := range config.SecLayerCtx.Handle2WriteChan {
		fmt.Println("=====", res)
		err := sendToRedis(res)
		if err != nil {
			log.Printf("{Result: %v} send to redis failed, with error: %v", res, err)
			continue
		}
	}
}

func sendToRedis(res *model.SecResult) (err error) {
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("JSON Marshal failed, with error: %v", err)
		return
	}

	fmt.Printf("Push Result to queue: %v", conf.Redis.Layer2proxyQueueName)
	conn := conf.Redis.RedisConn
	err = conn.LPush(conf.Redis.Layer2proxyQueueName, string(data)).Err()
	fmt.Println("Push over.")
	if err != nil {
		log.Printf("Lpush layer to proxy redis queue failed, with error: %v", err)
		return
	}
	log.Printf("{Result: %v} Lpush layer to proxy success.", string(data))
	return
}
