package models

import "time"

type TypeTransactionsTotal struct {
	DateMode        string
	StartDate       time.Time
	EndDate         time.Time
	Count           int
	SumSales        float64
	SumSalesCash    float64
	SumSalesCard    float64
	SumSalesOther   float64
	SumSalesMixed   float64
	SumReturns      float64
	SumReturnsCash  float64
	SumReturnsCard  float64
	SumReturnsOther float64
	SumReturnsMixed float64
	SumCash         float64
	SumCard         float64
	SumOther        float64
	SumMixed        float64
	Sum             float64
	Transactions    []TransactionEntity
	KassaTotal      []TypeKassaTotal
}

type TypeKassaTotal struct {
	KassaId          int64
	NameKassa        string
	NameOrganization string
	Count            int
	SumSales         float64
	SumSalesCash     float64
	SumSalesCard     float64
	SumSalesOther    float64
	SumSalesMixed    float64
	SumReturns       float64
	SumReturnsCash   float64
	SumReturnsCard   float64
	SumReturnsOther  float64
	SumReturnsMixed  float64
	SumCash          float64
	SumCard          float64
	SumOther         float64
	SumMixed         float64
	Sum              float64
}

type GoodElement struct {
	Name  string
	Size  string
	Price float64
	Qnt   int
}
