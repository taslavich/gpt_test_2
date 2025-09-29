package geoBadIp

import "errors"

var (
	BadIpFormatError = errors.New(
		"Bad IP Format",
	)

	InnerLookupIpError = errors.New(
		"Got error while looking up IP",
	)

	TorExitError = errors.New(
		"Tor exit IP",
	)

	PublicProxyError = errors.New(
		"Public Proxy IP",
	)

	AnonymousIpError = errors.New(
		"Anonymous IP",
	)
)
