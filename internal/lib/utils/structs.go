package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

func PrintAsJSON(data interface{}) (*[]byte, error) {
	//    var err := error
	p, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &p, nil
}

// вывести структуру в строку
// cummulative - выводить ли кумулятивную сумму
// count - выводить ли количество
func StructToString(data []models.Simple, cummulative bool, count bool) string {
	var b strings.Builder

	cummulativeSum := 0.0

	for _, s := range data {
		itemString := s.Item
		if itemString == "" {
			itemString = "неизвестно"
		}
		resultRow := ""

		if cummulative {
			cummulativeSum += s.Sum
			resultRow = "   " + itemString + ": " + FormatNumber(s.Sum) + " всего: " + FormatNumber(cummulativeSum)
		} else {
			resultRow = "   " + itemString + ": " + FormatNumber(s.Sum)
		}
		if count {
			resultRow += " кол-во: " + FormatNumber(float64(s.Count))
		}

		b.WriteString(resultRow + "\n")

	}
	return strings.TrimRight(b.String(), "\n")
}

func GroupByItem(arr []models.Simple) []models.Simple {
	m := make(map[string]models.Simple)

	for _, s := range arr {
		if agg, ok := m[s.Item]; ok {
			agg.Sum += s.Sum
			agg.Count += s.Count
			m[s.Item] = agg
		} else {
			m[s.Item] = s
		}
	}

	result := make([]models.Simple, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
