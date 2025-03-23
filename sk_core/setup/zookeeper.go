package setup

import (
	"encoding/json"
	"log"
	"time"

	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/samuel/go-zookeeper/zk"
)

func InitZk() {
	var hosts = []string{"localhost:2181"}
	option := zk.WithEventCallback(waitSecProductEvent)
	conn, _, err := zk.Connect(hosts, time.Second*5, option)
	if err != nil {
		log.Fatalf("Connect zookeeper failed, error: %v", err)
	}

	conf.Zk.ZkConn = conn
	conf.Zk.SecProductKey = "/product"

	go loadSecConf(conn)
}

func waitSecProductEvent(event zk.Event) {
	if event.Path == conf.Zk.SecProductKey {

	}
}

func loadSecConf(conn *zk.Conn) {
	log.Printf("Connect zookeeper success with produck key: %v", conf.Zk.SecProductKey)
	v, _, err := conn.Get(conf.Zk.SecProductKey)
	if err != nil {
		log.Printf("Get product info from zookeeper failed, error: %v", err)
		return
	}
	log.Printf("Get product info from zookeeper success.")

	var secProductInfo []*conf.SecProductInfoConf
	err = json.Unmarshal(v, &secProductInfo)
	if err != nil {
		log.Printf("JSON Unmsharl product info failed, error: %v", err)
	}

	updateSecProductInfo(secProductInfo)
}

func updateSecProductInfo(secProductInfo []*conf.SecProductInfoConf) {
	tmp := make(map[int]*conf.SecProductInfoConf, 1024)
	for _, v := range secProductInfo {
		tmp[v.ProductId] = v
	}
	conf.SecKill.RWBlackLock.Lock()
	conf.SecKill.SecProductInfoMap = tmp
	conf.SecKill.RWBlackLock.Unlock()
}
