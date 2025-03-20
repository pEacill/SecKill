package discovery

import (
	"fmt"
	"log"
)

type DiscoveryType string

const (
	ConsulDiscovery DiscoveryType = "Consul"
	EtcdDistcovery  DiscoveryType = "Etcd"
)

func CreateDiscoveryClient(discoveryType DiscoveryType, host, port string) (DiscoveryClient, error) {
	var client DiscoveryClient

	switch discoveryType {
	case ConsulDiscovery:
		client = NewConsulDiscoveryClient(host, port)
	case EtcdDistcovery:
		client = NewEtcdDiscoveryClient(host, port)
	default:
		log.Printf("Not support this discovery type.")
		return nil, fmt.Errorf("Not support this discovery type: %s.", discoveryType)
	}

	if client == nil {
		return nil, fmt.Errorf("Create Discovery Client fail.")
	}

	return client, nil
}
