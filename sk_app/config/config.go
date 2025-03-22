package config

import (
	"os"
	"sync"

	"github.com/go-kit/log"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/pEacill/SecKill/pkg/bootstrap"
	conf "github.com/pEacill/SecKill/pkg/config"
	"github.com/pEacill/SecKill/sk_app/model"
	"github.com/spf13/viper"
)

type SkAppCtx struct {
	SecReqChan       chan *model.SecRequest
	SecReqChanSize   int
	RWSecProductLock sync.RWMutex

	UserConnMap     map[string]chan *model.SecResult
	UserConnMapLock sync.Mutex
}

const (
	ProductStatusNormal       int = 1
	ProductStatusSaleOut      int = 2
	ProductStatusForceSaleOut int = 3
)

var SkAppContext = &SkAppCtx{
	UserConnMap: make(map[string]chan *model.SecResult),
	SecReqChan:  make(chan *model.SecRequest),
}

const kConfigType = "CONFIG_TYPE"

var ZipkinTracer *zipkin.Tracer
var Logger log.Logger

func init() {
	initLogger()
	initViper()
	// Logger.Log("Zipkin URL: ", url)
	// initTracer(url)
}

func initViper() {
	viper.AutomaticEnv()
	initViperDefault()

	if err := conf.LoadRemoteConfig(); err != nil {
		Logger.Log("Load Remote Config Fail!", err)
	}

	if err := conf.Sub("mysql", &conf.MysqlConfig); err != nil {
		conf.Logger.Log("Fail to Parse Mysql Config From Remote Config File!", err)
	}

	if err := conf.Sub("service", &conf.SecKill); err != nil {
		Logger.Log("Fail to parse service", err)
	}

	// Logger.Log("service: %v", conf.SecKill.CoreReadRedisGoroutineNum)

	if err := conf.Sub("redis", &conf.Redis); err != nil {
		Logger.Log("Fail to parse redis", err)
	}
	Logger.Log("service: %v", conf.Redis.Db)
	// if err := conf.Sub("trace", &conf.TraceConfig); err != nil {
	// 	Logger.Log("Fail to Parse Trace Config From Remote Config File!", err)
	// }

	// zipkinUrl := "http://" + conf.TraceConfig.Host + ":" + conf.TraceConfig.Port + conf.TraceConfig.Url
	// return zipkinUrl
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
