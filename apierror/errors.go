package apierror

import "errors"

var (
	ErrVerification          = errors.New("verification error")
	ErrLimitEmailCodeReached = errors.New("limit email code reached")
)
