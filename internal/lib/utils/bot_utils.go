package utils

import (
	"slices"
	"strconv"
	"strings"

	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

func ConvertTransToTotal(result modelsI.TypeTransactionsTotal, transactions []modelsI.TransactionEntity) modelsI.TypeTransactionsTotal {
	//SumSales := 0.0
	SumSalesCash := 0.0
	SumSalesCard := 0.0
	SumSalesOther := 0.0
	SumSalesMixed := 0.0
	//SumReturns := 0.0
	SumReturnsCash := 0.0
	SumReturnsCard := 0.0
	SumReturnsOther := 0.0
	SumReturnsMixed := 0.0
	// SumCash := 0.0
	// SumCard := 0.0
	// SumOther := 0.0
	// SumMixed := 0.0
	// Sum := 0.0

	count := 0

	for _, transaction := range transactions {
		if transaction.Type_operation == 1 { // продажа или возврат
			if transaction.Subtype.Valid && transaction.Subtype.Int64 == 2 { // продажа
				if transaction.Sum_operation.Valid {
					if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) == 1 {
						count++
						switch (*transaction.Paymenttypes)[0] {
						case 0:
							SumSalesCash += transaction.Sum_operation.Float64
						case 1:
							SumSalesCard += transaction.Sum_operation.Float64
						default:
							SumSalesOther += transaction.Sum_operation.Float64
						}
					}
					// если несколько типов платежей, то это смешанно
					if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) > 1 {
						count++
						SumSalesMixed += transaction.Sum_operation.Float64
					}

				}
			}
			if transaction.Subtype.Valid && transaction.Subtype.Int64 == 3 { // возврат
				if transaction.Sum_operation.Valid {
					if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) == 1 {
						//for _, paymentType := range *transaction.Paymenttypes {
						count++
						switch (*transaction.Paymenttypes)[0] {
						case 0:
							SumReturnsCash += transaction.Sum_operation.Float64
						case 1:
							SumReturnsCard += transaction.Sum_operation.Float64
						default:
							SumReturnsOther += transaction.Sum_operation.Float64
						}
						//}
					}
					// если несколько типов платежей, то это смешанно
					if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) > 1 {
						count++
						SumReturnsMixed += transaction.Sum_operation.Float64
					}
				}
			}
		}
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
		case 1:
			return "Выемка"
		}
	default:
		return "Неизвестно"
	}
	return "Неизвестно"
}
