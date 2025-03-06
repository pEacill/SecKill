package loadbalance

import (
	"errors"
	"math/rand"

	"github.com/pEacill/SecKill/pkg/common"
)

type RandomeLoadBalance struct{}

func (loadBalance *RandomeLoadBalance) SelectService(services []*common.ServiceInstance) (*common.ServiceInstance, error) {

	if services == nil || len(services) == 0 {
		return nil, errors.New("service instances are not exist")
	}

	return services[rand.Intn(len(services))], nil
}
