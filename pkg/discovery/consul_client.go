package discovery

import (
	"log"
	"strconv"

	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/pEacill/SecKill/pkg/common"
)

func NewDiscoveryClient(consulHost, consulPort string) *DiscoveryClientInstance {
	port, _ := strconv.Atoi(consulPort)

	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulHost + ":" + consulPort

	log.Printf("Initializing Consul API client with address: %s", consulConfig.Address)

	apiClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Printf("Failed to create Consul API client: %v", err)
		return nil
	}
	log.Println("Consul API client created successfully")

	client := consul.NewClient(apiClient)
	if client == nil {
		log.Printf("Failed to initialize Consul client")
		return nil
	}
	log.Println("Consul client initialized successfully")

	return &DiscoveryClientInstance{
		Host:   consulHost,
		Port:   port,
		config: consulConfig,
		client: client,
	}
}

func (d *DiscoveryClientInstance) Register(instanceId, svcHost, healthCheckURL, svcPort, svcName string, weight int, meta map[string]string, tags []string, logger *log.Logger) bool {
	if d.client == nil {
		if logger != nil {
			logger.Println("Consul client is not initialized")
		}
		return false
	}
	port, _ := strconv.Atoi(svcPort)
	logger.Println("service instance weight: ", weight)

	serviceRegistration := &api.AgentServiceRegistration{
		ID:      instanceId,
		Name:    svcName,
		Address: svcHost,
		Port:    port,
		Meta:    meta,
		Tags:    tags,
		Weights: &api.AgentWeights{
			Passing: weight,
		},
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           "http://" + svcHost + ":" + svcPort + healthCheckURL,
			Interval:                       "15s",
		},
	}

	err := d.client.Register(serviceRegistration)

	if err != nil {
		if logger != nil {
			logger.Println("Register Service Error!", err)
		}
		return false
	}
	if logger != nil {
		logger.Println("Register Service Success!")
	}
	return true
}

func (d *DiscoveryClientInstance) DeRegister(instanceId string, logger *log.Logger) bool {
	serviceRegistion := &api.AgentServiceRegistration{
		ID: instanceId,
	}

	err := d.client.Deregister(serviceRegistion)
	if err != nil {
		if logger != nil {
			logger.Println("Deregister Service Error!", err)
		}
		return false
	}

	if logger != nil {
		logger.Println("Deregister Service Success!")
	}

	return true
}

func newServiceInstance(service *api.AgentService) *common.ServiceInstance {
	rpcPort := service.Port - 1

	if service.Meta != nil {
		if rpcPortString, ok := service.Meta["rpcPort"]; ok {
			rpcPort, _ = strconv.Atoi(rpcPortString)
		}
	}

	return &common.ServiceInstance{
		Host:     service.Address,
		Port:     service.Port,
		GrpcPort: rpcPort,
		Weight:   service.Weights.Passing,
	}
}

func (d *DiscoveryClientInstance) DiscoverServices(serviceName string, logger *log.Logger) []*common.ServiceInstance {
	instanceList, ok := d.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*common.ServiceInstance)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	instanceList, ok = d.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*common.ServiceInstance)
	} else {
		go func() {
			params := make(map[string]interface{})
			params["type"] = "service"
			params["service"] = serviceName
			plan, _ := watch.Parse(params)
			plan.Handler = func(u uint64, i interface{}) {
				if i == nil {
					return
				}

				v, ok := i.([]*api.ServiceEntry)
				if !ok {
					return
				}

				if len(v) == 0 {
					d.instancesMap.Store(serviceName, []*common.ServiceInstance{})
				}

				var healthServices []*common.ServiceInstance

				for _, service := range v {
					if service.Checks.AggregatedStatus() == api.HealthPassing {
						healthServices = append(healthServices, newServiceInstance(service.Service))
					}
				}
				d.instancesMap.Store(serviceName, healthServices)
			}
			defer plan.Stop()
			plan.Run(d.config.Address)
		}()
	}

	entries, _, err := d.client.Service(serviceName, "", false, nil)
	if err != nil {
		d.instancesMap.Store(serviceName, []*common.ServiceInstance{})
		if logger != nil {
			logger.Println("Discover Service Error!", err)
		}
		return nil
	}
	instances := make([]*common.ServiceInstance, len(entries))
	for i := 0; i < len(instances); i++ {
		instances[i] = newServiceInstance(entries[i].Service)
	}
	d.instancesMap.Store(serviceName, instances)
	return instances
}
