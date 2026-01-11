package utils

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode"

	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/guregu/null/v5"
)

func GetGoodsFromCheque(data string) (modelsI.ChequeJSONList, error) {
	dataArr := strings.Split(data, "\n")
	var names modelsI.ChequeJSONList

	startGoodsBlock := 0
	endGoodsBlock := 0
	totalIndex := 0

	// получаем границы блока с товарами
	for index, line := range dataArr {
		if strings.HasPrefix(line, "*****") {
			startGoodsBlock = index + 1
		}
		if strings.HasPrefix(line, "-----") {
			if endGoodsBlock == 0 { // если найден второй раз, то пропускаем
				endGoodsBlock = index - 1
			}
		}
		if strings.Contains(line, "ИТОГО:") {
			totalIndex = index
		}
	}
	if startGoodsBlock == 0 || endGoodsBlock == 0 || totalIndex == 0 {
		return names, fmt.Errorf("не удалось найти блок с товарами в чеке")
	}

	var goodsBlocks []string
	currentBlock := ""

	// делим блок с товарами на отдельные блоки по 1 товару
	for i := startGoodsBlock; i <= endGoodsBlock; i++ {
		line := dataArr[i]
		if strings.Contains(line, "GTIN") || strings.Contains(line, "NTIN") {
			continue
		}
		if strings.HasSuffix(strings.TrimSpace(line), "₸") {
			currentBlock = currentBlock + line
			goodsBlocks = append(goodsBlocks, currentBlock)
			currentBlock = ""
		} else {
			currentBlock = currentBlock + line
		}
	}

	for _, block := range goodsBlocks {
		trimmedName, findedSize := trimNameSize(block)
		positionSum := strings.Index(block, "₸")
		price, err := trimPrice(block, positionSum)
		if err != nil {
			return names, fmt.Errorf("ошибка при парсинге цены товара: %v", err)
		}
		qnt := 1 // по умолчанию количество 1
		if !strings.Contains(block, "/СКИДКА") && !strings.Contains(block, "/НАЦЕНКА") {
			//fmt.Println("block: ", block)
			qnt, err = trimQnt(block)
			if err != nil {
				return names, fmt.Errorf("ошибка при парсинге количества товара: %v", err)
			}
		}
		if strings.Contains(block, "/СКИДКА") {
			// если это про скидку, то не добавляем строку в массив, меняем цену
			index := slices.IndexFunc(names, func(el modelsI.ChequeJSONElement) bool {
				// сравниваем по имени и размеру, если есть
				// внутри имени убираем пробелы, так как КОФД в секции скидки сам убирает пробелы
				return strings.ReplaceAll(el.Name, " ", "") == strings.ReplaceAll(trimmedName, " ", "") &&
					el.Size.String == findedSize
			})
			// fmt.Println("block", block)
			// fmt.Println("price", price)
			// fmt.Println("positionSum", positionSum)
			// fmt.Println("index", index)
			// fmt.Println("findedSize", findedSize)

			if index != -1 {
				//fmt.Println("NominalPrice", names[index].NominalPrice)
				discountPrice := names[index].NominalPrice - math.Round(price*100)/100
				names[index].DiscountPrice = discountPrice
				names[index].Sum = math.Round(discountPrice*100*float64(names[index].Qnt)) / 100
			}
		} else if strings.Contains(block, "/НАЦЕНКА") {
			index := slices.IndexFunc(names, func(el modelsI.ChequeJSONElement) bool {
				return el.Name == trimmedName
			})
			if index != -1 {
				discountPrice := names[index].NominalPrice + math.Round(price*100)/100
				names[index].DiscountPrice = discountPrice
				names[index].Sum = math.Round(discountPrice*100*float64(names[index].Qnt)) / 100
			}

		} else {
			names = append(names, modelsI.ChequeJSONElement{
				Name:          trimmedName,
				Size:          null.NewString(findedSize, findedSize != ""),
				NominalPrice:  math.Round(price*100) / 100,
				DiscountPrice: math.Round(price*100) / 100, // если нет скидки, то цена равна номинальной
				Qnt:           qnt,
				Sum:           math.Round(price*100*float64(qnt)) / 100,
			})
		}
	}
	var arraySum float64
	for item := range names {
		arraySum += math.Round(names[item].Sum*100) / 100
	}
	// считаем сумму всех товаров в чеке и сравниваем с итоговой суммой чека
	totalSum, err := getTotalSum(dataArr[totalIndex])
	if err != nil || totalSum != arraySum {
		fmt.Println("не совпадение суммы или ошибка при парсинге суммы: ", err)
		fmt.Println("чек: ", data)
	}

	//fmt.Printf("%+v\n", names)
	//fmt.Println("ИТОГО: ", totalSum)

	return names, nil
}

func trimNameSize(findedName string) (string, string) {
	trimmedName := ""
	findedSize := ""
	// если это строка со скидкой - убираем префикс
	findedName = strings.ReplaceAll(findedName, "ЖЕҢІЛДІК/СКИДКА", "")
	countBracket := strings.Count(findedName, "(")

	// ищем имя до первой левой скобки
	idx := strings.Index(findedName, "(")
	if idx != -1 {
		trimmedName = strings.TrimSpace(findedName[:idx])
	}

	// если 2 скобки, то ищем правую и размер между ними
	if countBracket >= 2 {
		idx2 := strings.Index(findedName, ")")
		if idx != -1 && idx2 != -1 {
			findedSize = strings.TrimSpace(findedName[idx+1 : idx2])
		}
	}
	return trimmedName, findedSize
}

func trimPrice(line string, positionSum int) (float64, error) {
	// ищем цену товара правее знака ₸
	priceString := ""
	var price float64

	for i := positionSum - 1; i >= 0; i-- {
		r := line[i]
		// если встретили разделитель - прекращаем
		if r == 'x' || r == '=' || r == ':' || r == ')' {
			priceString = line[i+1 : positionSum]
			break
		}
	}

	priceString = strings.TrimSpace(strings.ReplaceAll(priceString, "₸", ""))
	// ЧИСТИМ мусорные байты
	priceString = strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || r == '.' || r == ',' {
			return r
		}
		return -1
	}, priceString)

	priceString = strings.ReplaceAll(priceString, " ", "")  // убираем неразрывный пробел
	priceString = strings.ReplaceAll(priceString, ",", ".") // заменяем запятую на точку

	if priceString != "" {
		var err error
		price, err = ParseStringToFloat(priceString)
		if err != nil {
			return 0, fmt.Errorf("ошибка при парсинге цены товара: %v", err)
		}
	}
	return price, nil
}

func trimQnt(line string) (int, error) {
	//fmt.Println("line: ", line)
	var qnt int

	// не ищем если строка про скидку
	if strings.Contains(line, "/СКИДКА") {
		return 0, nil
	}
	if strings.Contains(line, "/НАЦЕНКА") {
		return 0, nil
	}

	idx := 0
	idx2 := 0
	qntString := ""

	// ищем справа налево открытую и закрытую скобку
	// из строки Cipo Туфли черный F6117MLB-25K-BLK (29) (шт)1 (Штука) x 26 900,00₸              = 26 900,00₸
	// а если Kapika Туфли летние синий 82157-2 (25) (шт)1 x 4 600,00₸
	//
	for i := len(line) - 1; i >= 0; i-- {
		if line[i] == '(' {
			idx = i
		}
		if line[i] == ')' && idx != 0 {
			idx2 = i
			break
		}
	}
	//fmt.Println("idx-1: ", idx-1, " idx2+1: ", idx2+1)

	if idx != 0 && idx2 != 0 {
		qntString = line[idx2+1 : idx]
	}

	// если получили пробел или пустое делаем попытку ориентироваться на строку между (шт) и x
	if qntString == "" || qntString == " " {
		start := strings.Index(line, "(шт)")
		end := strings.LastIndex(line, "x")

		if start != -1 && end != -1 && start < end {
			qntString = strings.TrimSpace(line[start+len("(шт)") : end])
			//fmt.Println("qntString2:", "|"+qntString+"|")
		}
	}
	// убираем знаки переноса строки
	qntString = strings.ReplaceAll(strings.ReplaceAll(qntString, "\n", ""), "\r", "")

	if qntString != "" {
		var err error
		//fmt.Println("qntString: ", "|"+qntString+"|")
		qnt, err = ParseStringToInt(qntString)
		if err != nil {
			return 0, fmt.Errorf("ошибка при парсинге количества товара: %v", err)
		}
	} else {
		return 0, fmt.Errorf("не удалось найти количество товара")
	}
	return qnt, nil
}

func getTotalSum(line string) (float64, error) {
	var totalSum float64
	var err error
	totalSumString := ""

	start := strings.Index(line, ":")
	end := strings.Index(line, "₸")
	if start != -1 && end != -1 && start < end {
		totalSumString = strings.TrimSpace(line[start+1 : end])
	}

	totalSum, err = ParseStringToFloat(totalSumString)
	if err != nil {
		return 0, fmt.Errorf("ошибка при парсинге суммы: %v", err)
	}
	return totalSum, nil
}
