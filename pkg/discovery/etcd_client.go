package discovery

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/pEacill/SecKill/pkg/common"
	"go.etcd.io/etcd/clientv3"
)

func NewEtcdDiscoveryClient(host, port string) *EtcdDistcoveryClientTnstance {
	port_int, _ := strconv.Atoi(port)

	etcdConfig := clientv3.Config{
		Endpoints:   []string{host + ":" + port},
		DialTimeout: 5 * time.Second,
	}
	log.Printf("Initializing Etcd API client with address: %s", (host + ":" + port))

	client, err := clientv3.New(etcdConfig)
	if err != nil {
		log.Printf("Create Etcd API client fail: %v", err)
		return nil
	}

	log.Printf("Create Etcd API client success.")

	return &EtcdDistcoveryClientTnstance{
		Host:   host,
		Port:   port_int,
		config: etcdConfig,
		client: client,
		leases: make(map[string]clientv3.LeaseID),
	}
}

func (e *EtcdDistcoveryClientTnstance) Register(instanceId, svcHost, healthCheckURL, svcPort, svcName string, weight int, meta map[string]string, tags []string, logger *log.Logger) bool {
	if e.client == nil {
		if logger != nil {
			logger.Println("Etcd client not init.")
		}
		return false
	}

	port, _ := strconv.Atoi(svcPort)
	if logger != nil {
		logger.Println("service instance weight: ", weight)
	}

	serviceInformation := map[string]interface{}{
		"ID":      instanceId,
		"Name":    svcName,
		"Address": svcHost,
		"Port":    port,
		"Meta":    meta,
		"Tags":    tags,
		"Weight":  weight,
		"Health":  "http://" + svcHost + ":" + svcPort + healthCheckURL,
	}
	serviceInformationJSON, err := json.Marshal(serviceInformation)
	if err != nil {
		if logger != nil {
			logger.Println("JSON service infomation error: %v.", err)
		}
		return false
	}

	ctx := context.Background()
	lease, err := e.client.Grant(ctx, 30)
	if err != nil {
		if logger != nil {
			logger.Println("Lease Create error: %v.", err)
		}
		return false
	}
	e.leases[instanceId] = lease.ID

	key := "/services/" + svcName + "/" + instanceId
	_, err = e.client.Put(ctx, key, string(serviceInformationJSON), clientv3.WithLease(lease.ID))
	if err != nil {
		if logger != nil {
			logger.Println("Registe service error: %v.", err)
		}
		return false
	}

	keepAliveCh, err := e.client.KeepAlive(ctx, lease.ID)
	if err != nil {
		if logger != nil {
			logger.Println("Keep Alive error: %v.", err)
		}
		return false
	}

	go func() {
		for {
			_, ok := <-keepAliveCh
			if !ok {
				if logger != nil {
					logger.Println("Service Not Alive.")
				}
				return
			}
		}
	}()

	if logger != nil {
		logger.Println("Registe success.")
	}

	return true
}

func (e *EtcdDistcoveryClientTnstance) DeRegister(instanceId string, logger *log.Logger) bool {
	return false
}

func (e *EtcdDistcoveryClientTnstance) DiscoverServices(serviceName string, logger *log.Logger) []*common.ServiceInstance {
	return nil
}
