package utils

import "strings"

const (
	// TelegramMaxMessageLen - лимит текста sendMessage.
	TelegramMaxMessageLen = 4096
	// TelegramMaxCaptionLen - лимит подписи к фото/медиагруппе у ботов
	// (2048 доступны только Premium-пользователям, но не ботам).
	TelegramMaxCaptionLen = 1024

	// маркер, который дописывается к обрезанному тексту
	trimMarker = "\n…"
)

// TelegramLen возвращает длину строки так, как ее считает Telegram - в кодовых
// единицах UTF-16, а не в байтах и не в рунах. Кириллица лежит в BMP и занимает
// 1 единицу (то есть 4096 русских букв проходят целиком, хотя в UTF-8 это ~8 Кб),
// а символы вне BMP (эмодзи вроде 🔍) - 2 единицы, хотя руна у них одна.
func TelegramLen(s string) int {
	length := 0
	for _, r := range s {
		if r > 0xFFFF {
			length += 2
		} else {
			length++
		}
	}
	return length
}

// TrimToTelegramLimit обрезает текст до limit кодовых единиц UTF-16, отбрасывая
// строки с конца и дописывая маркер обрезки. Режем именно по строкам, чтобы не
// разорвать HTML-тег посередине - Telegram на битой разметке отвечает 400
// can't parse entities. Предполагается, что теги в тексте не переносятся
// на следующую строку.
func TrimToTelegramLimit(s string, limit int) string {
	if TelegramLen(s) <= limit {
		return s
	}
	budget := limit - TelegramLen(trimMarker)
	if budget <= 0 {
		return ""
	}

	var sb strings.Builder
	total := 0
	for i, line := range strings.Split(s, "\n") {
		add := TelegramLen(line)
		if i > 0 {
			add++ // сам перевод строки
		}
		if total+add > budget {
			break
		}
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(line)
		total += add
	}

	if sb.Len() == 0 {
		// одна строка длиннее лимита - режем по рунам и отбрасываем
		// незакрытый тег в хвосте, если он попал под нож
		return trimRunes(s, budget) + trimMarker
	}
	return sb.String() + trimMarker
}

func trimRunes(s string, limit int) string {
	total := 0
	end := 0
	for i, r := range s {
		size := 1
		if r > 0xFFFF {
			size = 2
		}
		if total+size > limit {
			break
		}
		total += size
		end = i + len(string(r))
	}
	cut := s[:end]
	if idx := strings.LastIndex(cut, "<"); idx != -1 && !strings.Contains(cut[idx:], ">") {
		cut = cut[:idx]
	}
	return cut
}
