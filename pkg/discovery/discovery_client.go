package discovery

import (
	"log"
	"sync"

	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/pEacill/SecKill/pkg/common"
)

type DiscoveryClientInstance struct {
	Host string
	Port int

	config *api.Config
	client consul.Client
	mutex  sync.Mutex

	instancesMap sync.Map
}

type DiscoveryClient interface {
	Register(instanceId, svcHost, healthCheckURL, svcPort, svcName string, weight int, meta map[string]string, tags []string, logger *log.Logger) bool

	DeRegister(instanceId string, logger *log.Logger) bool

	DiscoverServices(serviceName string, logger *log.Logger) []*common.ServiceInstance
}
