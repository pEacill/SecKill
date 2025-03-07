package config

type TraceConf struct {
	Host string
	Port string
	Url  string
}

var (
	TraceConfig TraceConf
)
