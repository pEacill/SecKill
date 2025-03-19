package discovery

import (
	"log"
	"sync"

	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/pEacill/SecKill/pkg/common"
	"go.etcd.io/etcd/clientv3"
)

type ConsulDiscoveryClientInstance struct {
	Host string
	Port int

	config *api.Config
	client consul.Client
	mutex  sync.Mutex

	instancesMap sync.Map
}

type EtcdDistcoveryClientTnstance struct {
	Host string
	Port int

	config clientv3.Config
	client *clientv3.Client
	mutex  sync.Mutex

	instancesMap sync.Map
	leases       map[string]clientv3.LeaseID
}

type DiscoveryClient interface {
	Register(instanceId, svcHost, healthCheckURL, svcPort, svcName string, weight int, meta map[string]string, tags []string, logger *log.Logger) bool

	DeRegister(instanceId string, logger *log.Logger) bool

	DiscoverServices(serviceName string, logger *log.Logger) []*common.ServiceInstance
}
