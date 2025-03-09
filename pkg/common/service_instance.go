package common

type ServiceInstance struct {
	Host          string
	Port          int
	Weight        int
	CurrentWeight int

	GrpcPort int
}
