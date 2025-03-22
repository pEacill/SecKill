package setup

import (
	"log"
	"time"

	"github.com/go-redis/redis"
	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/pEacill/SecKill/sk_app/service/srv_redis"
	"github.com/unknwon/com"
)

func InitRedis() {
	log.Printf("Init Redis: %v", conf.Redis.Db)
	client := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Host + ":6379",
		Password:     conf.Redis.Password,
		DB:           conf.Redis.Db,
		PoolSize:     100,
		MinIdleConns: 20,
		IdleTimeout:  300 * time.Second,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Printf("Connect Redis Failed, Error: %v", err)
		return
	}
	log.Printf("Redis Init success.")
	conf.Redis.RedisConn = client

	loadBlackList(client)

}

func loadBlackList(conn *redis.Client) {
	conf.SecKill.IPBlackMap = make(map[string]bool, 10000)
	conf.SecKill.IDBlackMap = make(map[int]bool, 10000)

	// log.Printf("ID Blacklist Hash Key: %s", conf.Redis.IdBlackListHash)
	// log.Printf("IP Blacklist Hash Key: %s", conf.Redis.IpBlackListHash)

	idList, err := conn.HGetAll(conf.Redis.IdBlackListHash).Result()
	if err != nil {
		log.Printf("HGetall Failed, Error: %v", err)
		return
	}
	if len(idList) == 0 {
		log.Println("ID blacklist is empty in Redis.")
	}

	for _, v := range idList {
		id, err := com.StrTo(v).Int()
		if err != nil {
			log.Printf("Invalid UserId {%v}", id)
			continue
		}
		conf.SecKill.IDBlackMap[id] = true
	}

	ipList, err := conn.HGetAll(conf.Redis.IpBlackListHash).Result()
	if err != nil {
		log.Printf("HGetall Failed, Error: %v", err)
		return
	}
	if len(ipList) == 0 {
		log.Println("IP blacklist is empty in Redis.")
	}

	for _, v := range ipList {
		conf.SecKill.IPBlackMap[v] = true
	}

	// go syncIdBlackList(conn)
	// go syncIpBlackList(conn)

	return
}

func initRedisProcess() {
	log.Printf("Init Redis Process: {%d} write goroutines,  {%d} read goroutines.", conf.SecKill.AppWriteToHandleGoroutineNum, conf.SecKill.AppReadFromHandleGoroutineNum)

	for i := 0; i < conf.SecKill.AppWriteToHandleGoroutineNum; i++ {
		go srv_redis.WriteHandle()
	}

	for i := 0; i < conf.SecKill.AppReadFromHandleGoroutineNum; i++ {
		go srv_redis.ReadHandle()
	}
}

func syncIdBlackList(conn *redis.Client) {
	for {
		idArr, err := conn.BRPop(time.Minute, conf.Redis.IdBlackListQueue).Result()
		if err != nil {
			log.Printf("BRPop id failed, Error: %v", err)
			continue
		}

		id, _ := com.StrTo(idArr[1]).Int()
		conf.SecKill.RWBlackLock.Lock()
		conf.SecKill.IDBlackMap[id] = true
		conf.SecKill.RWBlackLock.Unlock()
	}
}

func syncIpBlackList(conn *redis.Client) {
	var ipList []string
	lastTime := time.Now().Unix()

	for {
		ipArr, err := conn.BRPop(time.Minute, conf.Redis.IpBlackListQueue).Result()
		if err != nil {
			log.Printf("BRPop id failed, Error: %v", err)
			continue
		}

		ip := ipArr[1]
		curTime := time.Now().Unix()
		ipList = append(ipList, ip)

		if len(ipList) > 100 || curTime-lastTime > 5 {
			conf.SecKill.RWBlackLock.Lock()
			{
				for _, v := range ipList {
					conf.SecKill.IPBlackMap[v] = true
				}
			}
			conf.SecKill.RWBlackLock.Unlock()

			lastTime = curTime
			log.Printf("sync ip list from redis success, ip[%v]", ipList)
		}
	}
}
