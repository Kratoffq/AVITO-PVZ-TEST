package pvz

import "errors"

var (
	// ErrNotFound ошибка, когда ПВЗ не найден
	ErrNotFound = errors.New("pvz not found")

	// ErrInvalidCity ошибка, когда указан неверный город
	ErrInvalidCity = errors.New("invalid city")

	// ErrInvalidDateRange ошибка, когда указан неверный диапазон дат
	ErrInvalidDateRange = errors.New("invalid date range")

	// ErrInvalidPagination ошибка, когда указаны неверные параметры пагинации
	ErrInvalidPagination = errors.New("invalid pagination parameters")
)
