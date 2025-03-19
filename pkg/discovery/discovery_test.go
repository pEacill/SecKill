package discovery_test

import (
	"testing"

	"github.com/pEacill/SecKill/pkg/bootstrap"
	"github.com/pEacill/SecKill/pkg/discovery"
)

func TestDiscoveryInit(t *testing.T) {
	discovery.InitComponent()

	if discovery.DiscoverService == nil {
		t.Fatalf("Failed to initialize ConsulService")
	}
	t.Logf("ConsulService initialized successfully")

	if discovery.Logger == nil {
		t.Fatalf("Logger not initialized")
	}
	t.Logf("Logger initialized successfully")

	discovery.Register()
	t.Logf("Service registered successfully")

	serviceName := bootstrap.DiscoverConfig.ServiceName
	instance, err := discovery.DiscoveryService(serviceName)
	if err != nil {
		t.Fatalf("Failed to discover service %s: %v", serviceName, err)
	}

	t.Logf("Discovered service instance: %+v", instance)

	discovery.Deregister()
	t.Logf("Service deregistered successfully")
}
