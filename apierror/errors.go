package apierror

import "errors"

var (
	ErrVerification     = errors.New("verification error")
	ErrEmailCodeTooSoon = errors.New("email code rate limit exceeded")
)
