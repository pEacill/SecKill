package loadbalance

import "github.com/pEacill/SecKill/pkg/common"

type LoadBalance interface {
	SelectService(services []*common.ServiceInstance) (*common.ServiceInstance, error)
}
