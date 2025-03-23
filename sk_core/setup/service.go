package setup

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	register "github.com/pEacill/SecKill/pkg/discovery"
	"github.com/pEacill/SecKill/sk_core/service/srv_redis"
)

func RunService() {
	srv_redis.RunProcess()
	errChan := make(chan error)
	go func() {
		register.Register()
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	register.Deregister()
	fmt.Println("sk-core shut down with error: %v", error)
}
