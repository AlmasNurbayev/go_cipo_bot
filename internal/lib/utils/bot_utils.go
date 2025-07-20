package utils

import (
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

	for _, transaction := range transactions {
		if transaction.Type_operation == 1 { // продажа или возврат
			if transaction.Subtype.Valid && transaction.Subtype.Int64 == 2 { // продажа
				if transaction.Sum_operation.Valid {
					if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) > 0 {
						for _, paymentType := range *transaction.Paymenttypes {
							switch paymentType {
							case 0:
								SumSalesCash += transaction.Sum_operation.Float64
							case 1:
								SumSalesCard += transaction.Sum_operation.Float64
							case 2:
								SumSalesMixed += transaction.Sum_operation.Float64
							default:
								SumSalesOther += transaction.Sum_operation.Float64
							}
						}
					}

				}
			}
			if transaction.Subtype.Valid && transaction.Subtype.Int64 == 3 { // возврат
				if transaction.Sum_operation.Valid {
					if transaction.Paymenttypes != nil && len(*transaction.Paymenttypes) > 0 {
						for _, paymentType := range *transaction.Paymenttypes {
							switch paymentType {
							case 0:
								SumReturnsCash += transaction.Sum_operation.Float64
							case 1:
								SumReturnsCard += transaction.Sum_operation.Float64
							case 2:
								SumReturnsMixed += transaction.Sum_operation.Float64
							default:
								SumReturnsOther += transaction.Sum_operation.Float64
							}
						}
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

	return result
}
