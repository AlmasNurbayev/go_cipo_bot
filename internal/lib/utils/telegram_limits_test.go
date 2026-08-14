package utils

import (
	"strings"
	"testing"
)

func TestTelegramLen(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"латиница", "abc", 3},
		{"кириллица считается как 1 единица на символ", "привет", 6},
		{"тенге - BMP", "₸", 1},
		{"эмодзи вне BMP - 2 единицы", "🔍", 2},
		{"составное эмодзи из двух BMP-символов", "⚠️", 2},
		{"пусто", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TelegramLen(tt.in); got != tt.want {
				t.Errorf("TelegramLen(%q) = %d, ожидалось %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestTelegramLenCyrillicFitsFullLimit(t *testing.T) {
	// 4096 кириллических символов - это ~8 Кб в UTF-8, но лимит Telegram они
	// занимают ровно на 4096, то есть проходят целиком
	s := strings.Repeat("я", TelegramMaxMessageLen)
	if got := TelegramLen(s); got != TelegramMaxMessageLen {
		t.Errorf("TelegramLen = %d, ожидалось %d", got, TelegramMaxMessageLen)
	}
	if len(s) != TelegramMaxMessageLen*2 {
		t.Errorf("байт = %d, ожидалось %d", len(s), TelegramMaxMessageLen*2)
	}
}

func TestTrimToTelegramLimitNoTrimNeeded(t *testing.T) {
	s := "<b>чек</b>\n • товар\n"
	if got := TrimToTelegramLimit(s, 100); got != s {
		t.Errorf("текст в пределах лимита изменен: %q", got)
	}
}

func TestTrimToTelegramLimitCutsWholeLines(t *testing.T) {
	s := "<b>первая</b>\n<i>вторая</i>\n<u>третья</u>"
	got := TrimToTelegramLimit(s, 30)

	if TelegramLen(got) > 30 {
		t.Errorf("результат длиннее лимита: %d", TelegramLen(got))
	}
	if !strings.HasSuffix(got, trimMarker) {
		t.Errorf("нет маркера обрезки: %q", got)
	}
	// теги не должны быть разорваны - число открывающих и закрывающих совпадает
	if strings.Count(got, "<")%2 != 0 || strings.Count(got, "<") != strings.Count(got, ">") {
		t.Errorf("разметка разорвана: %q", got)
	}
	if !strings.Contains(got, "<b>первая</b>") {
		t.Errorf("первая строка потеряна: %q", got)
	}
}

func TestTrimToTelegramLimitSingleLongLine(t *testing.T) {
	s := strings.Repeat("я", 100)
	got := TrimToTelegramLimit(s, 20)
	if TelegramLen(got) > 20 {
		t.Errorf("результат длиннее лимита: %d", TelegramLen(got))
	}
	if !strings.HasSuffix(got, trimMarker) {
		t.Errorf("нет маркера обрезки: %q", got)
	}
}

func TestTrimToTelegramLimitDropsBrokenTag(t *testing.T) {
	// одна длинная строка, обрезка приходится на середину тега
	s := "яяяяяяяяяяяяяяя<b>жирный</b>"
	got := TrimToTelegramLimit(s, 20)
	if strings.Count(got, "<") != strings.Count(got, ">") {
		t.Errorf("оборванный тег остался: %q", got)
	}
}
