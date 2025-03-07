package config

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/pEacill/SecKill/pkg/bootstrap"
	"github.com/pEacill/SecKill/pkg/discovery"
	"github.com/spf13/viper"
)

const kConfigType = "CONFIG_TYPE"

var Logger log.Logger
var ZipkinTracer *zipkin.Tracer

func init() {
	initLogger()
	url := initViper()
	Logger.Log("Zipkin URL: ", url)
	initTracer(url)
}

func initViperDefault() {
	viper.SetDefault(kConfigType, "yaml")
}

func initLogger() {
	Logger = log.NewLogfmtLogger(os.Stderr)
	Logger = log.With(Logger, "Time Stamper", log.DefaultTimestampUTC)
	Logger = log.With(Logger, "Caller", log.DefaultCaller)
}

func initViper() string {
	viper.AutomaticEnv()
	initViperDefault()

	if err := LoadRemoteConfig(); err != nil {
		Logger.Log("Load Remote Config Fail!", err)
	}

	if err := Sub("trace", &TraceConfig); err != nil {
		Logger.Log("Fail to Parse Trace Config From Remote Config File!", err)
	}
	zipkinUrl := "http://" + TraceConfig.Host + ":" + TraceConfig.Port + TraceConfig.Url
	return zipkinUrl
}

func LoadRemoteConfig() (err error) {
	serviceInstance, err := discovery.DiscoveryService(bootstrap.ConfigServerConfig.Id)
	if err != nil || serviceInstance == nil {
		Logger.Log("Discovery Config Server Fail!")
		return
	}
	configServer := "http://" + serviceInstance.Host + ":" + strconv.Itoa(serviceInstance.Port)
	confAddr := fmt.Sprintf("%v/%v/%v-%v.%v",
		configServer, bootstrap.ConfigServerConfig.Label,
		bootstrap.DiscoverConfig.ServiceName, bootstrap.ConfigServerConfig.Profile,
		viper.Get(kConfigType))
	resp, err := http.Get(confAddr)
	if err != nil {
		Logger.Log("G	et Config File Fail!")
		return
	}
	defer resp.Body.Close()

	viper.SetConfigType(viper.GetString(kConfigType))
	if err = viper.ReadConfig(resp.Body); err != nil {
		Logger.Log("Read Config File Fail!")
		return
	}
	Logger.Log("Load Config From: ", confAddr, "Success!")
	return nil
}

func Sub(key string, value interface{}) error {
	Logger.Log("Prefix of config file: %v", key)
	sub := viper.Sub(key)
	sub.AutomaticEnv()
	sub.SetEnvPrefix(key)
	return sub.Unmarshal(value)
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
		Logger.Log("tracer", "Zipkin", "type", "Native", "URL", zipkinURL)
	}
}
