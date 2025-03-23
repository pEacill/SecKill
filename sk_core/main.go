package main

import "github.com/pEacill/SecKill/sk_core/setup"

func main() {
	setup.InitZk()
	setup.InitRedis()
	setup.RunService()
}
