package errors

import "errors"

var (
	ErrLimitExceed        = errors.New("Rate limit exceed!")
	ErrInstanceNotExisted = errors.New("Instance not exist!")
	ErrRPCService         = errors.New("No this rpc service")
)
