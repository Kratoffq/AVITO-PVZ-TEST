package pvz

import "errors"

var (
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)
