package utils

import (
	"strconv"
	"strings"
	"unicode"
)

func ParseStringToFloat(s string) (float64, error) {
	// Удаляем все пробелы, включая неразрывные
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)

	// Заменяем запятую на точку
	cleaned = strings.ReplaceAll(cleaned, ",", ".")

	// Парсим как float
	return strconv.ParseFloat(cleaned, 64)
}

func ParseStringToInt(s string) (int, error) {
	// Удаляем все пробелы, включая неразрывные
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)

	// Заменяем запятую на точку
	cleaned = strings.ReplaceAll(cleaned, ",", ".")
	result, err := strconv.ParseInt(cleaned, 10, 64)

	return int(result), err
}
