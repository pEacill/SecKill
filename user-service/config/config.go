package config

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/pEacill/SecKill/pkg/bootstrap"
	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/spf13/viper"
)

const kConfigType = "CONFIG_TYPE"

var ZipkinTracer *zipkin.Tracer
var Logger log.Logger

func init() {
	initLogger()
	url := initViper()
	Logger.Log("Zipkin URL: ", url)
	initTracer(url)
}

func initViper() string {
	viper.AutomaticEnv()
	initViperDefault()

	if err := conf.LoadRemoteConfig(); err != nil {
		Logger.Log("Load Remote Config Fail!", err)
	}

	if err := conf.Sub("mysql", &conf.MysqlConfig); err != nil {
		conf.Logger.Log("Fail to Parse Mysql Config From Remote Config File!", err)
	}

	if err := conf.Sub("trace", &conf.TraceConfig); err != nil {
		Logger.Log("Fail to Parse Trace Config From Remote Config File!", err)
	}
	zipkinUrl := "http://" + conf.TraceConfig.Host + ":" + conf.TraceConfig.Port + conf.TraceConfig.Url
	return zipkinUrl
}

func initLogger() {
	Logger = log.NewLogfmtLogger(os.Stderr)
	Logger = log.With(Logger, "Time Stamper", log.DefaultTimestampUTC)
	Logger = log.With(Logger, "Caller", log.DefaultCaller)
}

func initViperDefault() {
	viper.SetDefault(kConfigType, "yaml")
}

func initTracer(zipkinURL string) {
	useNoopTracer := (zipkinURL == "")
	reporter := zipkinhttp.NewReporter(zipkinURL)
	var err error

	zEP, _ := zipkin.NewEndpoint(bootstrap.DiscoverConfig.ServiceName, bootstrap.HttpConfig.Port)
	ZipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer))
	if err != nil {
		Logger.Log("Get Zipkin Tracer Error:", err)
		os.Exit(1)
	}

	if !useNoopTracer {
		Logger.Log(
			"tracer", "Zipkin",
			"type", "Native",
			"URL", zipkinURL,
		)
	}
}
