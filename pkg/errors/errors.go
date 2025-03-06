package errors

import "errors"

var (
	ErrLimitExceed = errors.New("Rate limit exceed!")
)
