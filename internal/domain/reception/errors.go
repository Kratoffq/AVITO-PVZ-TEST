package reception

import "errors"

var (
	// ErrNotFound возвращается, когда приёмка не найдена
	ErrNotFound = errors.New("reception not found")

	// ErrNoOpenReception возвращается, когда нет открытой приёмки
	ErrNoOpenReception = errors.New("no open reception")

	// ErrReceptionAlreadyOpen возвращается, когда для ПВЗ уже есть открытая приёмка
	ErrReceptionAlreadyOpen = errors.New("reception already open")
)
