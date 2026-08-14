package utils

import (
	"math"
	"testing"
)

// чек с двумя одинаковыми товарами - на каждый КОФД печатает свою строку скидки
const chequeTwoSameGoods = `   Incore
              БСН/БИН 800727301256

                    Продажа
Чектің реттік нөмірі/Порядковый номер чека 10034
Ауысым/Смена №1291
ФИСКАЛДЫҚ БЕЛГІ/ФИСКАЛЬНЫЙ ПРИЗНАК: 850132799055
                       КАССИР КОДЫ/КОД КАССИРА 1
УАҚЫТЫ/ВРЕМЯ: 14.08.2026 18:47:23
КЗН/ЗНМ SWK00426032                КСН/ИНК 33812
КТН/РНМ 010102355028
***********************************************
Cipo розовый 3514SA-FL-PNK (33) (шт)
1 (Штука) x 37 800,00₸              = 37 800,00₸
GTIN: 2200000083517
Cipo розовый 3514SA-FL-PNK (33) (шт)
1 (Штука) x 37 800,00₸              = 37 800,00₸
GTIN: 2200000083517
Cipo синий 9950-37-40 (36) (шт)
1 (Штука) x 28 500,00₸              = 28 500,00₸
GTIN: 2200000055606
NTIN: 0200423769328
ЖЕҢІЛДІК/СКИДКА
Cipo розовый 3514SA-FL-PNK (33) (шт)  12 474,00₸
ЖЕҢІЛДІК/СКИДКА
Cipo розовый 3514SA-FL-PNK (33) (шт)  12 474,00₸
ЖЕҢІЛДІК/СКИДКА
Cipo синий 9950-37-40 (36) (шт)        9 405,00₸
------------------------------------------------
Төленген сома/Сумма оплаты                 0,00₸
    Банковская карта:                 69 747,00₸
Қайтарым сомасы/Сумма сдачи                0,00₸
Жеңілдік сомасы/Сумма скидки          34 353,00₸
үстеме сомасы/Сумма наценки                0,00₸
ҚҚС сомасы/Сумма НДС                       0,00₸
БАРЛЫҒЫ/ИТОГО:69 747,00₸
------------------------------------------------
`

// чек с одним товаром со скидкой
const chequeOneGoodWithDiscount = `   Incore
              БСН/БИН 800727301256
                    Продажа
УАҚЫТЫ/ВРЕМЯ: 14.08.2026 18:47:23
***********************************************
Cipo синий 9950-37-40 (36) (шт)
1 (Штука) x 28 500,00₸              = 28 500,00₸
GTIN: 2200000055606
ЖЕҢІЛДІК/СКИДКА
Cipo синий 9950-37-40 (36) (шт)        9 405,00₸
------------------------------------------------
    Банковская карта:                 19 095,00₸
БАРЛЫҒЫ/ИТОГО:19 095,00₸
------------------------------------------------
`

// чек без скидки, количество больше одного
const chequeWithoutDiscount = `   Incore
              БСН/БИН 800727301256
                    Продажа
УАҚЫТЫ/ВРЕМЯ: 14.08.2026 18:47:23
***********************************************
Cipo синий 9950-37-40 (36) (шт)
2 (Штука) x 28 500,00₸              = 57 000,00₸
GTIN: 2200000055606
------------------------------------------------
    Банковская карта:                 57 000,00₸
БАРЛЫҒЫ/ИТОГО:57 000,00₸
------------------------------------------------
`

type expectedGood struct {
	name          string
	size          string
	nominalPrice  float64
	discountPrice float64
	qnt           int
	sum           float64
}

func TestGetGoodsFromCheque(t *testing.T) {
	tests := []struct {
		name     string
		cheque   string
		expected []expectedGood
		totalSum float64
	}{
		{
			name:   "два одинаковых товара со скидкой - скидка применяется к обоим",
			cheque: chequeTwoSameGoods,
			expected: []expectedGood{
				{"Cipo розовый 3514SA-FL-PNK", "33", 37800, 25326, 1, 25326},
				{"Cipo розовый 3514SA-FL-PNK", "33", 37800, 25326, 1, 25326},
				{"Cipo синий 9950-37-40", "36", 28500, 19095, 1, 19095},
			},
			totalSum: 69747,
		},
		{
			name:   "один товар со скидкой",
			cheque: chequeOneGoodWithDiscount,
			expected: []expectedGood{
				{"Cipo синий 9950-37-40", "36", 28500, 19095, 1, 19095},
			},
			totalSum: 19095,
		},
		{
			name:   "товар без скидки - цена со скидкой равна номинальной",
			cheque: chequeWithoutDiscount,
			expected: []expectedGood{
				{"Cipo синий 9950-37-40", "36", 28500, 28500, 2, 57000},
			},
			totalSum: 57000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names, err := GetGoodsFromCheque(tt.cheque)
			if err != nil {
				t.Fatalf("неожиданная ошибка: %v", err)
			}
			if len(names) != len(tt.expected) {
				t.Fatalf("получено %d позиций, ожидалось %d: %+v", len(names), len(tt.expected), names)
			}

			var arraySum float64
			for i, expected := range tt.expected {
				got := names[i]
				if got.Name != expected.name {
					t.Errorf("позиция %d: имя %q, ожидалось %q", i, got.Name, expected.name)
				}
				if got.Size.String != expected.size {
					t.Errorf("позиция %d: размер %q, ожидался %q", i, got.Size.String, expected.size)
				}
				if got.NominalPrice != expected.nominalPrice {
					t.Errorf("позиция %d: номинальная цена %.2f, ожидалась %.2f", i, got.NominalPrice, expected.nominalPrice)
				}
				if got.DiscountPrice != expected.discountPrice {
					t.Errorf("позиция %d: цена со скидкой %.2f, ожидалась %.2f", i, got.DiscountPrice, expected.discountPrice)
				}
				if got.Qnt != expected.qnt {
					t.Errorf("позиция %d: количество %d, ожидалось %d", i, got.Qnt, expected.qnt)
				}
				if got.Sum != expected.sum {
					t.Errorf("позиция %d: сумма %.2f, ожидалась %.2f", i, got.Sum, expected.sum)
				}
				arraySum += got.Sum
			}

			// сумма позиций должна сходиться с ИТОГО чека
			if math.Abs(arraySum-tt.totalSum) > 0.01 {
				t.Errorf("сумма позиций %.2f не совпадает с итогом чека %.2f", arraySum, tt.totalSum)
			}
		})
	}
}
