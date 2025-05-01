package product

import "errors"

var (
	// ErrNotFound возвращается, когда товар не найден
	ErrNotFound = errors.New("product not found")
)
