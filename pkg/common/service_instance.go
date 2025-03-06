package common

type ServiceInstance struct {
	Host          string
	Port          string
	Weight        int
	CurrentWeight int

	GrpcPort int
}
