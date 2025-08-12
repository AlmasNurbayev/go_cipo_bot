package utils

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func ConvertTransToTotal(transactions []modelsI.TransactionEntity,
	ListKassas []modelsI.KassaEntity) modelsI.TypeTransactionsTotal {

	SumSalesCash := 0.0
	SumSalesCard := 0.0
	SumSalesOther := 0.0
	SumSalesMixed := 0.0
	SumReturnsCash := 0.0
	SumReturnsCard := 0.0
	SumReturnsOther := 0.0
	SumReturnsMixed := 0.0
	SumInputCash := 0.0
	SumOutputCash := 0.0
	// SumCash := 0.0
	// SumCard := 0.0
	// SumOther := 0.0
	// SumMixed := 0.0
	// Sum := 0.0

	count := 0

	var result modelsI.TypeTransactionsTotal
	var kassaTotal []modelsI.TypeKassaTotal

	for _, kassa := range ListKassas {
		kassaSumSalesCash := 0.0
		kassaSumSalesCard := 0.0
		kassaSumSalesOther := 0.0
		kassaSumSalesMixed := 0.0
		kassaSumReturnsCash := 0.0
		kassaSumReturnsCard := 0.0
		kassaSumReturnsOther := 0.0
		kassaSumReturnsMixed := 0.0
		kassaSumInputCash := 0.0
		kassaSumOutputCash := 0.0
		kassaCount := 0
		for _, transaction := range transactions {
			if transaction.Kassa_id != kassa.Id {
				continue
			}
			if transaction.Type_operation == 6 { // выемка или внесение
				if transaction.Subtype.Valid && transaction.Subtype.Int64 == 0 { // внесение
					SumInputCash += transaction.Sum_operation.Float64
					kassaSumInputCash += transaction.Sum_operation.Float64
				}
				if transaction.Subtype.Valid && transaction.Subtype.Int64 == 1 { // выемка
					SumOutputCash += transaction.Sum_operation.Float64
					kassaSumOutputCash += transaction.Sum_operation.Float64
				}
			}
			if transaction.Type_operation == 1 { // продажа или возврат
				if transaction.Subtype.Valid && transaction.Subtype.Int64 == 2 { // продажа
					if transaction.Sum_operation.Valid {
						if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) == 1 {
							count++
							kassaCount++
							switch (*transaction.Paymenttypes)[0] {
							case 0:
								SumSalesCash += transaction.Sum_operation.Float64
								kassaSumSalesCash += transaction.Sum_operation.Float64
							case 1:
								SumSalesCard += transaction.Sum_operation.Float64
								kassaSumSalesCard += transaction.Sum_operation.Float64
							default:
								SumSalesOther += transaction.Sum_operation.Float64
								kassaSumSalesOther += transaction.Sum_operation.Float64
							}
						}
						// если несколько типов платежей, то это смешанно
						if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) > 1 {
							count++
							kassaCount++
							SumSalesMixed += transaction.Sum_operation.Float64
							kassaSumSalesMixed += transaction.Sum_operation.Float64
						}
					}
				}
				if transaction.Subtype.Valid && transaction.Subtype.Int64 == 3 { // возврат
					if transaction.Sum_operation.Valid {
						if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) == 1 {
							//for _, paymentType := range *transaction.Paymenttypes {
							count++
							kassaCount++
							switch (*transaction.Paymenttypes)[0] {
							case 0:
								SumReturnsCash += transaction.Sum_operation.Float64
								kassaSumReturnsCash += transaction.Sum_operation.Float64
							case 1:
								SumReturnsCard += transaction.Sum_operation.Float64
								kassaSumReturnsCard += transaction.Sum_operation.Float64
							default:
								SumReturnsOther += transaction.Sum_operation.Float64
								kassaSumReturnsOther += transaction.Sum_operation.Float64
							}
							//}
						}
						// если несколько типов платежей, то это смешанно
						if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) > 1 {
							count++
							kassaCount++
							SumReturnsMixed += transaction.Sum_operation.Float64
							kassaSumReturnsMixed += transaction.Sum_operation.Float64
						}
					}
				}
			}
		}
		if kassaCount == 0 {
			continue // если нет чеков по кассе, то пропускаем
		}
		kassaTotal = append(kassaTotal, modelsI.TypeKassaTotal{
			KassaId:          kassa.Id,
			NameKassa:        kassa.Name_kassa,
			NameOrganization: kassa.Organization_name.String,
			Count:            kassaCount,
			SumSales:         kassaSumSalesCash + kassaSumSalesCard + kassaSumSalesOther + kassaSumSalesMixed,
			SumSalesCash:     kassaSumSalesCash,
			SumSalesCard:     kassaSumSalesCard,
			SumSalesOther:    kassaSumSalesOther,
			SumSalesMixed:    kassaSumSalesMixed,
			SumReturns:       kassaSumReturnsCard + kassaSumReturnsCash + kassaSumReturnsOther + kassaSumReturnsMixed,
			SumReturnsCash:   kassaSumReturnsCash,
			SumReturnsCard:   kassaSumReturnsCard,
			SumReturnsOther:  kassaSumReturnsOther,
			SumReturnsMixed:  kassaSumReturnsMixed,
			SumInputCash:     kassaSumInputCash,
			SumOutputCash:    kassaSumOutputCash,
			SumCash:          kassaSumSalesCash - kassaSumReturnsCash,
			SumCard:          kassaSumSalesCard - kassaSumReturnsCard,
			SumOther:         kassaSumSalesOther - kassaSumReturnsOther,
			SumMixed:         kassaSumSalesMixed - kassaSumReturnsMixed,
			Sum: kassaSumSalesCash - kassaSumReturnsCash + kassaSumSalesCard - kassaSumReturnsCard +
				kassaSumSalesOther - kassaSumReturnsOther + kassaSumSalesMixed - kassaSumReturnsMixed,
		})

	}

	result.SumSales = SumSalesCash + SumSalesCard + SumSalesOther + SumSalesMixed
	result.SumSalesCash = SumSalesCash
	result.SumSalesCard = SumSalesCard
	result.SumSalesOther = SumSalesOther
	result.SumSalesMixed = SumSalesMixed
	result.SumReturns = SumReturnsCard + SumReturnsCash + SumReturnsOther + SumReturnsMixed
	result.SumReturnsCash = SumReturnsCash
	result.SumReturnsCard = SumReturnsCard
	result.SumReturnsOther = SumReturnsOther
	result.SumReturnsMixed = SumReturnsMixed
	result.SumCash = SumSalesCash - SumReturnsCash
	result.SumCard = SumSalesCard - SumReturnsCard
	result.SumOther = SumSalesOther - SumReturnsOther
	result.SumMixed = SumSalesMixed - SumReturnsMixed
	result.Sum = result.SumCash + result.SumCard + result.SumOther + result.SumMixed
	result.Count = count
	result.SumInputCash = SumInputCash
	result.SumOutputCash = SumOutputCash
	result.KassaTotal = kassaTotal
	return result
}

func ConvertNewOperationToMessageText(message modelsI.MessagesType,
	kassas []modelsI.KassaEntity) string {
	var sb strings.Builder
	for _, tx := range message.Transactions {
		sumStr := "<b>" + FormatNumber(tx.Sum_operation.Float64) + "</b>"

		kassa := slices.IndexFunc(kassas, func(k modelsI.KassaEntity) bool { return k.Id == tx.Kassa_id })
		kassaString := ""

		if kassa != -1 {
			kassaString = kassas[kassa].Name_kassa
		}

		sb.WriteString(" 💸 " + kassaString + " №" + strconv.FormatInt(tx.Id, 10) + " от " +
			tx.Operationdate.Time.Format("15:04") + " " +
			GetTypeOperationText(tx) +
			" " + sumStr + "\n")

		for _, item := range tx.ChequeJSON {
			sb.WriteString(" • " + item.Name + " (" + item.Size.String + ") ₸ " + FormatNumber(item.Sum) + "\n")
		}

	}
	return sb.String()
}

func GetTypeOperationText(oper modelsI.TransactionEntity) string {
	switch oper.Type_operation {
	case 1:
		switch oper.Subtype.Int64 {
		case 2:
			return "Продажа"
		case 3:
			return "Возврат"
		}
	case 2:
		switch oper.Subtype.Int64 {
		case 0:
			return "Закрытие смены"
		}
	case 6:
		switch oper.Subtype.Int64 {
		case 0:
			return "Внесение"
		case 1:
			return "Выемка"
		}
	default:
		return "Неизвестно"
	}
	return "Неизвестно"
}

func SendAction(chatID int64, action string, b *bot.Bot) error {
	// Отправляем действие, чтобы не было "пустого" сообщения
	var typeAction models.ChatAction
	switch action {
	case "typing":
		{
			typeAction = models.ChatActionTyping
		}
	case "upload_photo":
		{
			typeAction = models.ChatActionUploadPhoto
		}
	case "upload_document":
		{
			typeAction = models.ChatActionUploadDocument
		}
	default:
		{
			return errors.New("неизвестное действие: " + action)
		}

	}
	_, err := b.SendChatAction(context.Background(), &bot.SendChatActionParams{
		ChatID: chatID,
		Action: typeAction,
	})
	if err != nil {
		return err
	}
	// Задержка чтобы статус был виден
	time.Sleep(500 * time.Millisecond)
	return nil
}
