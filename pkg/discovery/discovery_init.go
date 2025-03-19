package discovery

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pEacill/SecKill/pkg/bootstrap"
	"github.com/pEacill/SecKill/pkg/common"
	"github.com/pEacill/SecKill/pkg/errors"
	"github.com/pEacill/SecKill/pkg/loadbalance"
	uuid "github.com/satori/go.uuid"
)

var DiscoverService DiscoveryClient
var LoadBalance loadbalance.LoadBalance
var Logger *log.Logger

func init() {
	InitComponent()
}

func InitComponent() {
	DiscoverService = NewConsulDiscoveryClient(bootstrap.DiscoverConfig.Host, bootstrap.DiscoverConfig.Port)
	LoadBalance = new(loadbalance.RandomeLoadBalance)
	Logger = log.New(os.Stderr, "", log.LstdFlags)
}

func Register() {
	if DiscoverService == nil {
		panic(0)
	}

	instanceId := bootstrap.DiscoverConfig.InstanceId
	if instanceId == "" {
		instanceId = bootstrap.DiscoverConfig.ServiceName + uuid.NewV4().String()
	}

	if !DiscoverService.Register(
		instanceId,
		bootstrap.HttpConfig.Host,
		"/health",
		bootstrap.HttpConfig.Port,
		bootstrap.DiscoverConfig.ServiceName,
		bootstrap.DiscoverConfig.Weight,
		map[string]string{
			"rpcPort": bootstrap.RpcConfig.Port,
		},
		nil,
		Logger,
	) {
		Logger.Printf("register service %s failed. ", bootstrap.DiscoverConfig.ServiceName)
		panic(0)
	}
	Logger.Printf("%s-service for service %s success.", bootstrap.DiscoverConfig.ServiceName, bootstrap.DiscoverConfig.ServiceName)
}

func Deregister() {
	if DiscoverService == nil {
		panic(0)
	}

	instanceId := bootstrap.DiscoverConfig.InstanceId
	if instanceId == "" {
		instanceId = bootstrap.DiscoverConfig.ServiceName + "-" + uuid.NewV4().String()
	}

	if !DiscoverService.DeRegister(instanceId, Logger) {
		Logger.Printf("deregister for service %s failed.", bootstrap.DiscoverConfig.ServiceName)
		panic(0)
	}
}

func DiscoveryService(serviceName string) (*common.ServiceInstance, error) {
	instances := DiscoverService.DiscoverServices(serviceName, Logger)
	if len(instances) < 1 {
		Logger.Printf("no available client for %s.", serviceName)
		return nil, errors.ErrInstanceNotExisted
	}
	return LoadBalance.SelectService(instances)
}

func CheckHealth(w http.ResponseWriter, r *http.Request) {
	Logger.Println("Health Check!")
	_, err := fmt.Fprintf(w, "Server is OK!")
	if err != nil {
		Logger.Println(err)
	}
}
