package endpoint

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/pEacill/SecKill/oauth-service/errors"
	"github.com/pEacill/SecKill/oauth-service/model"
	"github.com/pEacill/SecKill/oauth-service/service"
)

const (
	OAuth2DetailsKey       = "OAuth2Details"
	OAuth2ClientDetailsKey = "OAuth2ClientDetails"
	OAuth2ErrorKey         = "OAuth2Error"
)

type OAuth2Endpoints struct {
	TokenEndpoint          endpoint.Endpoint
	CheckTokenEndpoint     endpoint.Endpoint
	GRPCCheckTokenEndpoint endpoint.Endpoint
	HealthCheckEndpoint    endpoint.Endpoint
}

// Verify whether the context of the request contains client information.
func MakeClientAuthorizationMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if err, ok := ctx.Value(OAuth2ErrorKey).(error); ok {
				return nil, err
			}
			if _, ok := ctx.Value(OAuth2ClientDetailsKey).(*model.ClientDetails); !ok {
				return nil, errors.ErrInvalidClientRequest
			}

			return next(ctx, request)
		}
	}
}

// Verify whether the context of the request contains OAuth2Details information.
func MakeOAuth2AuthorizationMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if err, ok := ctx.Value(OAuth2ErrorKey).(error); ok {
				return nil, err
			}
			if _, ok := ctx.Value(OAuth2DetailsKey).(*model.OAuth2Details); !ok {
				return nil, errors.ErrInvalidUserRequest
			}
			return next(ctx, request)
		}
	}
}

// Verify whether the user has permission to access.
func MakeAuthorityAuthorizationMiddleware(authority string, logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {

		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			if err, ok := ctx.Value(OAuth2ErrorKey).(error); ok {
				return nil, err
			}
			if details, ok := ctx.Value(OAuth2DetailsKey).(*model.OAuth2Details); !ok {
				return nil, errors.ErrInvalidClientRequest
			} else {
				for _, value := range details.User.Autorities {
					if value == authority {
						return next(ctx, request)
					}
				}
				return nil, errors.ErrNotPermit
			}
		}
	}
}

type TokenRequest struct {
	GrantType string
	Reader    *http.Request
}

type TokenResponse struct {
	AccessToken *model.OAuth2Token `json:"access_token"`
	Error       string             `json:"error"`
}

type CheckTokenRequest struct {
	Token         string
	ClientDetails model.ClientDetails
}

type ChecktokenResponse struct {
	Oauth2Details *model.OAuth2Details `json:"o_auth_details"`
	Error         string               `json:"error"`
}

type HealthCheckRequest struct{}

type HealthCheckResponse struct {
	Status bool `json:"status"`
}

type SimpleRequest struct{}

type SimpleResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

type AdminRequest struct{}

type AdminResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

// Make Endpoint
func MakeTokenEndpoint(svc service.TokenGranter, clientService service.ClientDetailsService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*TokenRequest)
		token, err := svc.Grant(ctx, req.GrantType, ctx.Value(OAuth2ClientDetailsKey).(*model.ClientDetails), req.Reader)
		var errString = ""
		if err != nil {
			errString = err.Error()
		}

		return TokenResponse{
			AccessToken: token,
			Error:       errString,
		}, nil
	}
}

func MakeCheckTokenEndpoint(svc service.TokenService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*CheckTokenRequest)
		tokenDetails, err := svc.GetOAuth2DetailsByAccessToken(req.Token)

		var errString = ""
		if err != nil {
			errString = err.Error()
		}

		return ChecktokenResponse{
			Oauth2Details: tokenDetails,
			Error:         errString,
		}, nil
	}
}

func MakeHealthCheckEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthCheckResponse{
			Status: status,
		}, nil
	}
}
