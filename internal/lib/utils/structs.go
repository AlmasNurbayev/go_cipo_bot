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

func StructToString(data []models.Simple) string {
	var b strings.Builder

	for _, s := range data {
		itemString := s.Item
		if itemString == "" {
			itemString = "неизвестно"
		}
		b.WriteString("  " + itemString + ": " + FormatNumber(s.Sum) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func GroupByItem(arr []models.Simple) []models.Simple {
	m := make(map[string]float64)
	for _, s := range arr {
		m[s.Item] += s.Sum
	}

	result := make([]models.Simple, 0, len(m))
	for item, sum := range m {
		result = append(result, models.Simple{Item: item, Sum: sum})
	}
	return result
}
