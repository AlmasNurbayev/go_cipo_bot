package utils

import (
	"fmt"
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

func FormatNumber(n float64) string {
	s := fmt.Sprintf("%.0f", n) // округляем и убираем дробную часть
	var result []string
	for i, c := range reverse(s) {
		if i > 0 && i%3 == 0 {
			result = append(result, " ")
		}
		result = append(result, string(c))
	}
	return reverse(strings.Join(result, ""))
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
