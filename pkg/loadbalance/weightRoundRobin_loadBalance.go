package loadbalance

import (
	"errors"

	"github.com/pEacill/SecKill/pkg/common"
)

type WeightRoundRobinLoadBalance struct{}

func (loadBalance *WeightRoundRobinLoadBalance) SelectService(services []*common.ServiceInstance) (best *common.ServiceInstance, err error) {

	if services == nil || len(services) == 0 {
		return nil, errors.New("service instances are not exist")
	}

	total := 0
	for i := 0; i < len(services); i++ {
		w := services[i]
		if w == nil {
			continue
		}

		w.CurrentWeight += w.Weight

		total += w.Weight
		if best == nil || w.CurrentWeight > best.CurrentWeight {
			best = w
		}
	}

	if best == nil {
		return nil, nil
	}

	best.CurrentWeight -= total
	return best, nil
}
