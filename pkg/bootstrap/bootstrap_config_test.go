package bootstrap

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestBootStrapConfig(t *testing.T) {
	viper.SetConfigName("bootstrap")
	viper.AddConfigPath("./")
	viper.SetConfigType("yaml")

	assert.NotPanics(t, func() {
		InitLocalViper()
	}, "InitializeConfig should not panic")

	assert.NotEmpty(t, HttpConfig.Port, "HTTP config should be loaded")
	assert.NotEmpty(t, DiscoverConfig.ServiceName, "Discover config should be loaded")
	assert.NotEmpty(t, ConfigServerConfig.Id, "Config server Id should be loaded")
	assert.NotEmpty(t, RpcConfig.Port, "RPC config should be loaded")
}
