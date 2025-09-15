package ping

import "errors"

var (
	errClosed   = errors.New("pinger closed")
	errNotBound = errors.New("need at least one bind address")
	errNotCIDR  = errors.New("empty cidr addresses")
	errNotHOSTS = errors.New("empty host addresses")
)

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }
