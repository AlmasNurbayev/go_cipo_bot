package other

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type LogEntry struct {
	Date            string `json:"date"`
	Status          string `json:"status"`
	BasePrefix      string `json:"basePrefix"`
	CountQnt        int    `json:"countQnt"`
	IsContainImages bool   `json:"isContainImages"`
	NameFile        string `json:"nameFile"`
}

// RawLogLine вспомогательная структура для парсинга JSON строки лога
type RawLogLine struct {
	Msg        string `json:"msg"`
	Time       string `json:"time"`
	DateSchema string `json:"date_schema"`
	NameLog    string `json:"name_log"`
	BasePrefix string `json:"base_prefix"`
	TotalQnt   int    `json:"TotalQnt"`
	Op         string `json:"op"`
}

func ParseAppLog(filePath string, lineLimit int) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка: файла нет или нет доступа: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Находим позицию, с которой начинаются последние N строк
	offset, err := getLastLinesOffset(file, lineLimit)
	if err != nil {
		return nil, err
	}

	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	var results []LogEntry
	var currentEntry *LogEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var line RawLogLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue
		}

		msgLower := strings.ToLower(line.Msg)

		// 1. Начало секции
		if strings.Contains(msgLower, "==== init parserjson") {
			currentEntry = &LogEntry{}
			continue
		}

		// 2. Данные регистратора
		if line.Msg == "New registrator:" && line.Op == "parserJSON.parserRegistrator" {
			if currentEntry != nil && len(line.DateSchema) >= 16 {
				currentEntry.Date = line.DateSchema[:16]
				currentEntry.BasePrefix = line.BasePrefix
				currentEntry.CountQnt = line.TotalQnt
				currentEntry.NameFile = line.NameLog
			}
		}

		// успешный импорт картинок
		if strings.Contains(msgLower, "images exists and copied successfully") {
			if currentEntry != nil {
				currentEntry.IsContainImages = true
			}
		}

		// 3. Успех[cite: 1]
		if strings.Contains(msgLower, "success finished") {
			if currentEntry != nil {
				currentEntry.Status = "success"
				results = append(results, *currentEntry)
				currentEntry = nil
			}
		}

		// 4. Ошибка[cite: 1]
		if strings.Contains(msgLower, "error finished") {
			if currentEntry != nil {
				currentEntry.Status = "error"
				if len(line.Time) >= 16 {
					currentEntry.Date = line.Time[:16]
				}
				results = append(results, *currentEntry)
				currentEntry = nil
			}
		}
	}

	// Разворачиваем для получения порядка "от поздних к ранним"
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results, nil
}

// Вспомогательная функция для поиска смещения последних N строк
func getLastLinesOffset(file *os.File, limit int) (int64, error) {
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	size := stat.Size()
	var cursor int64 = 0
	cnt := 0
	buffer := make([]byte, 1024)

	for cursor < size && cnt <= limit {
		cursor += 1024
		if cursor > size {
			cursor = size
		}

		_, err := file.Seek(-cursor, io.SeekEnd)
		if err != nil {
			return 0, err
		}

		n, _ := file.Read(buffer)
		for i := n - 1; i >= 0; i-- {
			if buffer[i] == '\n' {
				cnt++
				if cnt > limit {
					return size - cursor + int64(i) + 1, nil
				}
			}
		}
	}
	return 0, nil // Если строк меньше лимита, читаем с начала
}
