package core

import "errors"

// Ошибки библиотеки.
var (
	ErrInvalidParams = errors.New("некорректные параметры")
	ErrDivByZero     = errors.New("делитель содержит ноль")
	ErrEmpty         = errors.New("пустая величина или набор")
	ErrIndex         = errors.New("индекс вне диапазона")
)
