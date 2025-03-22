package setup

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/samuel/go-zookeeper/zk"
)

func InitZookeeper() {
	var hosts = []string{"localhost:2181"}
	// option := zk.WithEventCallback(waitSecProductEvent)
	conn, _, err := zk.Connect(hosts, time.Second*5)
	if err != nil {
		fmt.Printf("Connect zookeeper error: %v", err)
		return
	}

	conf.Zk.ZkConn = conn
	conf.Zk.SecProductKey = "/product"

}

func loadSecConf(conn *zk.Conn) {
	log.Printf("Connect zk success: %s", conf.Zk.SecProductKey)
	v, _, err := conn.Get(conf.Zk.SecProductKey)
	if err != nil {
		log.Printf("Get product info failed, error: %v", err)
		return
	}
	log.Printf("Get product info")
	var secProductInfo []*conf.SecProductInfoConf
	err = json.Unmarshal(v, &secProductInfo)
	if err != nil {
		log.Printf("Unmsharl second product info failed, error: %v", err)
	}
	updateSecProductInfo(secProductInfo)
}

func waitSecProductEvent(event zk.Event) {
	log.Print(">>>>>>>>>>>>>>>>>>>")
	log.Println("path:", event.Path)
	log.Println("type:", event.Type.String())
	log.Println("state:", event.State.String())
	log.Println("<<<<<<<<<<<<<<<<<<<")
	if event.Path == conf.Zk.SecProductKey {
	}
}

func updateSecProductInfo(secProductInfo []*conf.SecProductInfoConf) {
	tmp := make(map[int]*conf.SecProductInfoConf, 1024)
	for _, v := range secProductInfo {
		log.Printf("Update SecProduct info: %v", v)
		tmp[v.ProductId] = v
	}
	conf.SecKill.RWBlackLock.Lock()
	conf.SecKill.SecProductInfoMap = tmp
	conf.SecKill.RWBlackLock.Unlock()
}
