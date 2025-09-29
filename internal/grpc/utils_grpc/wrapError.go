package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WrapError(prefix string, wrapError error) error {
	st, ok := status.FromError(wrapError)
	if !ok {
		return status.Errorf(
			codes.Unknown,
			"Because got unknown error %w: %w",
			prefix,
			st.Err(),
		)
	}
	return status.Errorf(
		st.Code(),
		"%w: %w",
		prefix,
		st.Err(),
	)
}
