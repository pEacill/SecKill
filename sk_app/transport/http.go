package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	gozipkin "github.com/openzipkin/zipkin-go"
	"github.com/pEacill/SecKill/sk_app/endpoint"
	"github.com/pEacill/SecKill/sk_app/model"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func decodeSecInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var secInfoRequest endpoint.SecInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&secInfoRequest); err != nil {
		return nil, err
	}
	return secInfoRequest, nil
}

func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpoint.HealthCheckRequest{}, nil
}

func decodeTestRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpoint.HealthCheckRequest{}, nil
}

func decodeSecInfoListRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpoint.SecInfoListRequest{}, nil
}

func decodeSecKillRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var secRequest model.SecRequest
	if err := json.NewDecoder(r.Body).Decode(&secRequest); err != nil {
		return nil, err
	}
	return secRequest, nil
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func MakeHttpHandler(ctx context.Context, endpoints endpoint.SkAppEndpoints, zipkinTracer *gozipkin.Tracer, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer, zipkin.Name("http-transport"))

	options := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
		zipkinServer,
	}

	r.Methods("POST").Path("/sec/info").Handler(kithttp.NewServer(
		endpoints.GetSecInfoEndpoint,
		decodeSecInfoRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/sec/list").Handler(kithttp.NewServer(
		endpoints.GetSecInfoListEndpoint,
		decodeSecInfoListRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/sec/kil").Handler(kithttp.NewServer(
		endpoints.SecKillEndpoint,
		decodeSecKillRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/sec/test").Handler(kithttp.NewServer(
		endpoints.TestEndpoint,
		decodeTestRequest,
		encodeResponse,
		options...,
	))

	r.Path("/metrics").Handler(promhttp.Handler())

	r.Methods("GET").Path("/health").Handler(kithttp.NewServer(
		endpoints.HealthCheckEndPoint,
		decodeHealthCheckRequest,
		encodeResponse,
		options...,
	))

	return r
}
