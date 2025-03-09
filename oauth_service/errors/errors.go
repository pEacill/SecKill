package errors

import "errors"

var (
	ErrNotSupportGrantType                   = errors.New("Grant type is not supported!")
	ErrInvalidUsernameOrPasswordRequest      = errors.New("Invalid username or password!")
	ErrInvalidUsernameOrPasswordRequestEmpty = errors.New("username or password is empty!")
	ErrInvalidTokenRequest                   = errors.New("Invalid token!")
	ErrInvalidUserInfo                       = errors.New("invalid user info")
	ErrNotSupportOperation                   = errors.New("No support operation")
	ErrClientMessage                         = errors.New("Invalid client")

	ErrorBadRequest         = errors.New("Invalid request parameter")
	ErrorGrantTypeRequest   = errors.New("Invalid request grant type")
	ErrorTokenRequest       = errors.New("Invalid request token")
	ErrInvalidClientRequest = errors.New("Invalid client message")
	ErrInvalidUserRequest   = errors.New("Invalid user message")
	ErrNotPermit            = errors.New("Not permit")

	ErrExpiredToken = errors.New("Token is expired")
)
