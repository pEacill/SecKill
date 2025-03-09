package client

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/pEacill/SecKill/pb"
)

func TestUserClientImpl_CheckUser(t *testing.T) {
	client, _ := NewUserClient("user", nil, genTracerAct(nil))

	if response, err := client.CheckUser(context.Background(), nil, &pb.UserRequest{
		Username: "testuser",
		Password: "testpassword",
	}); err == nil {
		fmt.Println(response.Result)
	} else {
		fmt.Println(err.Error())
	}
}

func genTracerAct(tracer opentracing.Tracer) opentracing.Tracer {
	if tracer != nil {
		return tracer
	}

	zipkinUrl := "http://localhost:9411/api/v2/spans"
	zipkinRecorder := "localhost:9000"

	reporter := zipkinhttp.NewReporter(zipkinUrl)

	localEndpoint, err := zipkin.NewEndpoint("user-client", zipkinRecorder)
	if err != nil {
		log.Fatalf("zipkin.NewEndpoint err: %v", err)
	}

	nativeTracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(localEndpoint),
		zipkin.WithSharedSpans(true),
	)
	if err != nil {
		log.Fatalf("zipkin.NewTracer err: %v", err)
	}

	opentracingTracer := zipkinot.Wrap(nativeTracer)

	return opentracingTracer
}
