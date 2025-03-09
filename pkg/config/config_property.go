package config

type TraceConf struct {
	Host string
	Port string
	Url  string
}

type MysqlConf struct {
	Host string
	Port string
	User string
	Pwd  string
	Db   string
}

var (
	TraceConfig TraceConf
	MysqlConfig MysqlConf
)
