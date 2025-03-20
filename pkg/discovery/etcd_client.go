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
	if e.client == nil {
		if logger != nil {
			logger.Println("Etcd client not init.")
		}
		return false
	}

	ctx := context.Background()

	if leaseId, exist := e.leases[instanceId]; exist {
		_, err := e.client.Revoke(ctx, leaseId)
		if err != nil {
			if logger != nil {
				logger.Println("Register fail (lease fail) %v.", err)
			}
			return false
		}
		delete(e.leases, instanceId)
	} else {
		resp, err := e.client.Get(ctx, "/services/", clientv3.WithPrefix())
		if err != nil {
			if logger != nil {
				logger.Println("Register fail (search service fail) %v.", err)
			}
			return false
		}

		for _, kv := range resp.Kvs {
			var serviceData map[string]interface{}

			if err := json.Unmarshal(kv.Value, &serviceData); err != nil {
				continue
			}

			if id, ok := serviceData["ID"].(string); ok && id == instanceId {
				_, err = e.client.Delete(ctx, string(kv.Key))
				if err != nil {
					if logger != nil {
						logger.Println("Register fail (delete service in Etcd fail) %v.", err)
					}
					return false
				}
				break
			}
		}
	}

	if logger != nil {
		logger.Println("Register services success.")
	}

	return true
}

func (e *EtcdDistcoveryClientTnstance) DiscoverServices(serviceName string, logger *log.Logger) []*common.ServiceInstance {
	instanceList, ok := e.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*common.ServiceInstance)
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	instanceList, ok = e.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*common.ServiceInstance)
	} else {
		go e.watchService(serviceName, logger)

		instances := e.getServiceInstances(serviceName, logger)
		e.instancesMap.Store(serviceName, instances)
		return instances
	}
}

func (e *EtcdDistcoveryClientTnstance) getServiceInstances(serviceName string, logger *log.Logger) []*common.ServiceInstance {
	ctx := context.Background()
	resp, err := e.client.Get(ctx, "/services/"+serviceName+"/", clientv3.WithPrefix())
	if err != nil {
		if logger != nil {
			logger.Println("discover services error: %v.", err)
		}
		return []*common.ServiceInstance{}
	}

	instances := make([]*common.ServiceInstance, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var serviceData map[string]interface{}
		if err := json.Unmarshal(kv.Value, &serviceData); err != nil {
			if logger != nil {
				logger.Println("discover services error(json unmarshal error): %v.", err)
			}
			continue
		}

		portValue, _ := serviceData["Port"].(float64)
		port := int(portValue)
		rpcPort := port - 1

		if meta, ok := serviceData["Meta"].(map[string]interface{}); ok {
			if rpcPortStr, ok := meta["rpcPort"].(string); ok {
				rpcPort, _ = strconv.Atoi(rpcPortStr)
			}
		}

		weightValue, _ := serviceData["Weight"].(float64)
		weight := int(weightValue)

		instance := &common.ServiceInstance{
			Host:     serviceData["Address"].(string),
			Port:     port,
			GrpcPort: rpcPort,
			Weight:   weight,
		}

		instances = append(instances, instance)
	}

	return instances
}

func (e *EtcdDistcoveryClientTnstance) watchService(serviceName string, logger *log.Logger) {
	prefix := "/services/" + serviceName + "/"
	ctx := context.Background()

	watchCh := e.client.Watch(ctx, prefix, clientv3.WithPrefix())

	for watchResp := range watchCh {
		if watchResp.Canceled {
			if logger != nil {
				logger.Println("Watch service over. service name: %v.", serviceName)
			}
			return
		}

		instances := e.getServiceInstances(serviceName, logger)
		e.instancesMap.Store(serviceName, instances)

		if logger != nil {
			logger.Println("Update services: %v, instances num: %v", serviceName, len(instances))
		}
	}
}
