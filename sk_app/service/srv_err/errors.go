package srv_err

import "errors"

const (
	SeckillSucc = 2002

	ErrRetry = 2001
)

const (
	ErrNotFoundProductId = 1001
	ErrActiveNotStart    = 1002
	ErrActiveAlreadyEnd  = 1003
	ErrActiveSaleOut     = 1004
	ErrUserServiceBusy   = 1005
	ErrProcessTimeout    = 1006
	ErrClientClosed      = 1007
)

var errMsg = map[int]string{
	ErrRetry:    "Please try again.",
	SeckillSucc: "SecKill success.",
}

func GetErrMsg(code int) error {
	return errors.New(errMsg[code])
}
