package models

import "time"

// описания колонок в values
// "4" - Id строки
// "Аренда Мухамедханова" - description (описание)
// "Аренда помещений" - cost_item (статья БДР)
// "Мухамедханова" - division (подразделение)
// "01.01.2024" - period (отчетный период)
// "30.12.2023" - date (дата оплаты)
// "400 000,00" - sum (сумма)
// "KZT" - currency (валюта)
// "" - sumUSD (сумма в USD)
// "Kaspi Pay" - account (счет)
// "ИП Incore" - organization (организация)
type GSheetsResponseV1 struct {
	Range          string     `json:"range"`          // '2024'!A4:X1533
	MajorDimension string     `json:"majorDimension"` // ROWS
	Values         [][]string `json:"values"`         // FORMATTED_VALUE
}

type GSheetsEntityV1 struct {
	Id           int       `json:"id" db:"id"`
	Description  string    `json:"description" db:"description"`
	Category     string    `json:"cost_item" db:"cost_item"`
	Division     string    `json:"division" db:"division"`
	Period       time.Time `json:"period" db:"period"`
	Date         time.Time `json:"date" db:"date"`
	Sum          float64   `json:"sum" db:"sum"`
	Currency     string    `json:"currency" db:"currency"`
	SumUSD       float64   `json:"sumUSD" db:"sumUSD"`
	Account      string    `json:"account" db:"account"`
	Organization string    `json:"organization" db:"organization"`
}
