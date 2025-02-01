package quik

import "strconv"

// Counter - структура для генерации последовательности чисел
type Counter struct {
	current int
}

// Next возвращает следующее число в последовательности
func (c *Counter) Next() string {
	c.current++
	return strconv.Itoa(c.current)
}

// NewCounter создает новый счетчик, начиная с указанного числа
func NewCounter(start int) *Counter {
	return &Counter{current: start - 1}
}
