package bootstrap

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func init() {
	InitLocalViper()
}

func InitLocalViper() {
	viper.AutomaticEnv()
	initBootstrapConfig()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("err:%s\n", err)
	}

	if err := subParse("http", &HttpConfig); err != nil {
		log.Fatal("Fail to parse Http config", err)
	}
	if err := subParse("discover", &DiscoverConfig); err != nil {
		log.Fatal("Fail to parse Discover config", err)
	}
	if err := subParse("config", &ConfigServerConfig); err != nil {
		log.Fatal("Fail to parse config server", err)
	}

	if err := subParse("rpc", &RpcConfig); err != nil {
		log.Fatal("Fail to parse rpc server", err)
	}
}

func initBootstrapConfig() {
	viper.SetConfigName("bootstrap")
	viper.AddConfigPath("./")
	viper.SetConfigType("yaml")
}

func subParse(key string, value interface{}) error {
	log.Printf("Prefix of config file: %v", key)
	sub := viper.Sub(key)
	sub.AutomaticEnv()
	sub.SetEnvPrefix(key)
	return sub.Unmarshal(value)
}
