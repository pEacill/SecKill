package main

import (
	"github.com/pEacill/SecKill/pkg/bootstrap"
	"github.com/pEacill/SecKill/sk_app/setup"
)

func main() {
	setup.InitZookeeper()
	setup.InitRedis()
	setup.InitServer(bootstrap.HttpConfig.Host, bootstrap.HttpConfig.Port)
}
