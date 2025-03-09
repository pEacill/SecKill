package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	zipkingo "github.com/openzipkin/zipkin-go"
	"github.com/pEacill/SecKill/oauth-service/endpoint"
	"github.com/pEacill/SecKill/oauth-service/errors"
	"github.com/pEacill/SecKill/oauth-service/service"
)

func decodeTokenRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	grantType := r.URL.Query().Get("grant_type")
	if grantType == "" {
		return nil, errors.ErrorGrantTypeRequest
	}

	return &endpoint.TokenRequest{
		GrantType: grantType,
		Reader:    r,
	}, nil
}

func decodeCheckTokenRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	tokenValue := r.URL.Query().Get("token")
	if tokenValue == "" {
		return nil, errors.ErrorTokenRequest
	}

	return &endpoint.CheckTokenRequest{
		Token: tokenValue,
	}, nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func encodeJsonResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpoint.HealthCheckRequest{}, nil
}

func MakeHttpHandler(ctx context.Context, endpoints endpoint.OAuth2Endpoints, tokenService service.TokenService,
	clientService service.ClientDetailsService, zipkinTracer *zipkingo.Tracer, logger log.Logger) http.Handler {

	r := mux.NewRouter()
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer, zipkin.Name("oaut-transport"))

	options := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
		zipkinServer,
	}

	// r.Path("/metrics").Handler(promhttp.Handler())

	ClientAuthorizationOptions := []kithttp.ServerOption{
		kithttp.ServerBefore(makeClientAuthorizationContext(clientService, logger)),
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
		zipkinServer,
	}

	r.Methods("POST").Path("/oauth/token").Handler(kithttp.NewServer(
		endpoints.TokenEndpoint,
		decodeTokenRequest,
		encodeJsonResponse,
		ClientAuthorizationOptions...,
	))

	r.Methods("POST").Path("/oauth/check_token").Handler(kithttp.NewServer(
		endpoints.CheckTokenEndpoint,
		decodeCheckTokenRequest,
		encodeJsonResponse,
		ClientAuthorizationOptions...,
	))

	r.Methods("GET").Path("/health").Handler(kithttp.NewServer(
		endpoints.HealthCheckEndpoint,
		decodeHealthCheckRequest,
		encodeJsonResponse,
		options...,
	))

	return r
}

// get client info in the http request and store it in the context
func makeClientAuthorizationContext(clientDetailsService service.ClientDetailsService, logger log.Logger) kithttp.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		if clientId, clientSecret, ok := r.BasicAuth(); ok {
			clientDetails, err := clientDetailsService.GetClientDetailsByClientId(ctx, clientId, clientSecret)
			if err == nil {
				return context.WithValue(ctx, endpoint.OAuth2ClientDetailsKey, clientDetails)
			}
		}
		return context.WithValue(ctx, endpoint.OAuth2ErrorKey, errors.ErrInvalidClientRequest)
	}
}

func makeOAuth2AuthorizationContext(tokenService service.TokenService, logger log.Logger) kithttp.RequestFunc {

	return func(ctx context.Context, r *http.Request) context.Context {

		accessTokenValue := r.Header.Get("Authorization")
		var err error
		if accessTokenValue != "" {
			oauth2Details, err := tokenService.GetOAuth2DetailsByAccessToken(accessTokenValue)
			if err == nil {
				return context.WithValue(ctx, endpoint.OAuth2DetailsKey, oauth2Details)
			}
		} else {
			err = errors.ErrorTokenRequest
		}

		return context.WithValue(ctx, endpoint.OAuth2ErrorKey, err)
	}
}
