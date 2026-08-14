package summary

import (
	"strconv"
	"strings"
	"testing"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
)

func testRows(n int) []checkRow {
	rows := make([]checkRow, 0, n)
	for i := 1; i <= n; i++ {
		rows = append(rows, checkRow{
			ID:       int64(1000 + i),
			Num:      strconv.Itoa(i),
			DateTime: "01.08.26 14:23",
			Type:     "Продажа",
			Sum:      "45 000",
			Card:     "45 000",
			Cash:     "—",
		})
	}
	return rows
}

func TestRenderChecksTable(t *testing.T) {
	table := renderChecksTable(testRows(2))
	lines := strings.Split(strings.TrimRight(table, "\n"), "\n")

	// шапка + разделитель + 2 строки
	if len(lines) != 4 {
		t.Fatalf("строк в таблице: %d, ожидалось 4\n%s", len(lines), table)
	}
	if !strings.Contains(lines[0], "в т.ч. карта") || !strings.Contains(lines[0], "в т.ч. нал") {
		t.Errorf("нет колонок оплаты в шапке: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "|") || !strings.Contains(lines[1], "---") {
		t.Errorf("нет строки выравнивания GFM: %q", lines[1])
	}

	// все строки одинаковой ширины - иначе моноширинный запасной вариант разъедется
	width := len([]rune(lines[0]))
	for i, line := range lines {
		if got := len([]rune(line)); got != width {
			t.Errorf("строка %d шириной %d, ожидалось %d:\n%s", i, got, width, table)
		}
	}
	// в каждой строке 6 колонок - 7 разделителей
	for i, line := range lines {
		if got := strings.Count(line, "|"); got != 7 {
			t.Errorf("строка %d: разделителей %d, ожидалось 7: %q", i, got, line)
		}
	}
	t.Log("\n" + table)
}

func TestRenderChecksTableEmpty(t *testing.T) {
	table := renderChecksTable(nil)
	if !strings.Contains(table, "№") {
		t.Errorf("шапка потеряна: %q", table)
	}
}

func TestSplitTableRowsByMaxRows(t *testing.T) {
	// 100 чеков - это 3 части, размер части ограничивают кнопки, а не длина текста
	parts := splitTableRows(testRows(100), richMessageMaxLen, maxRowsPerRichTable)
	if len(parts) != 3 {
		t.Fatalf("частей: %d, ожидалось 3", len(parts))
	}
	total := 0
	for _, part := range parts {
		if len(part) > maxRowsPerRichTable {
			t.Errorf("в части %d строк, лимит %d", len(part), maxRowsPerRichTable)
		}
		if got := utils.TelegramLen(renderChecksTable(part)); got > richMessageMaxLen {
			t.Errorf("часть длиннее лимита rich-сообщения: %d", got)
		}
		total += len(part)
	}
	if total != 100 {
		t.Errorf("строк после разбиения: %d, ожидалось 100", total)
	}
}

func TestTableKeyboard(t *testing.T) {
	markup := tableKeyboard(testRows(maxRowsPerRichTable))

	if len(markup.InlineKeyboard) != maxRowsPerRichTable/maxButtonsPerRow {
		t.Errorf("рядов кнопок: %d, ожидалось %d",
			len(markup.InlineKeyboard), maxRowsPerRichTable/maxButtonsPerRow)
	}

	count := 0
	for _, row := range markup.InlineKeyboard {
		if len(row) > maxButtonsPerRow {
			t.Errorf("в ряду %d кнопок, лимит %d", len(row), maxButtonsPerRow)
		}
		count += len(row)
	}
	if count != maxRowsPerRichTable {
		t.Errorf("кнопок: %d, ожидалось %d", count, maxRowsPerRichTable)
	}

	// номер на кнопке совпадает с колонкой "№", а данные ведут на чек
	first := markup.InlineKeyboard[0][0]
	if first.Text != "1" || first.CallbackData != "getCheck_1001" {
		t.Errorf("первая кнопка: %q / %q", first.Text, first.CallbackData)
	}
}

func TestTableKeyboardNotFullRow(t *testing.T) {
	markup := tableKeyboard(testRows(maxButtonsPerRow + 1))
	if len(markup.InlineKeyboard) != 2 {
		t.Fatalf("рядов: %d, ожидалось 2", len(markup.InlineKeyboard))
	}
	if len(markup.InlineKeyboard[1]) != 1 {
		t.Errorf("в последнем ряду %d кнопок, ожидалась 1", len(markup.InlineKeyboard[1]))
	}
}

func TestSplitTableRowsByLength(t *testing.T) {
	// запасной вариант режется по лимиту обычного сообщения
	rows := testRows(500)
	parts := splitTableRows(rows, utils.TelegramMaxMessageLen, 0)
	total := 0
	for i, part := range parts {
		if got := utils.TelegramLen(renderChecksTable(part)); got > utils.TelegramMaxMessageLen {
			t.Errorf("часть %d длиной %d, лимит %d", i, got, utils.TelegramMaxMessageLen)
		}
		total += len(part)
	}
	if total != len(rows) {
		t.Errorf("строк после разбиения: %d, ожидалось %d", total, len(rows))
	}
}
