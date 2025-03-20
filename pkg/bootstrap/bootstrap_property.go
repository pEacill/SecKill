package bootstrap

type HttpConf struct {
	Host string
	Port string
}

type RpcConf struct {
	Port string
}

type DiscoverConf struct {
	Type        string
	Host        string
	Port        string
	ServiceName string
	Weight      int
	InstanceId  string
}

type ConfigServerConf struct {
	Id      string
	Profile string
	Label   string
}

var (
	HttpConfig         HttpConf
	RpcConfig          RpcConf
	DiscoverConfig     DiscoverConf
	ConfigServerConfig ConfigServerConf
)
