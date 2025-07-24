package utils

import (
	"fmt"
	"strings"

	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

func GetGoodsFromCheque(data string) ([]modelsI.GoodElement, error) {
	dataArr := strings.Split(data, "\n")
	var names []modelsI.GoodElement

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

	fmt.Println(startGoodsBlock, endGoodsBlock, totalIndex)
	//var existedIndexes []int
	firstindex, lastindex := 0, 0

	//firstIndex := 0
	for i := startGoodsBlock; i <= endGoodsBlock; i++ {
		line := dataArr[i]

		//fmt.Println("line", line)

		// получаем границы отдельных товаров
		positionSum := strings.Index(line, "₸")
		if positionSum == -1 {
			if firstindex == 0 {
				firstindex = i
			}
		} else {
			lastindex = i - 1
			fmt.Println("firstindex", firstindex, "lastindex", lastindex)
			findedName := ""
			trimmedName := ""
			findedSize := ""
			// получаем название товара вместе с размером и ед изм
			if firstindex == lastindex {
				findedName = dataArr[firstindex]
			} else {
				findedName = strings.Join(dataArr[firstindex:lastindex+1], "")
			}
			if strings.Contains(findedName, "СКИДКА") { // пропускаем скидки
				continue
			}
			// обрезаем строку до 2 открывающих скобок
			countBracket := strings.Count(findedName, "(")
			if countBracket >= 2 {
				idx := strings.Index(findedName, "(")
				idx2 := strings.Index(findedName, ")")
				if idx != -1 {
					trimmedName = strings.TrimSpace(findedName[:idx])
				}
				if idx != -1 && idx2 != -1 {
					findedSize = strings.TrimSpace(findedName[idx+1 : idx2])
				}
			}
			// ищем цену товара
			priceString := ""
			var price float64
			positionX := strings.Index(line, " x ")
			if positionX != -1 {
				priceString = strings.TrimSpace(line[positionX+3 : positionSum])
			}
			if priceString != "" {
				var err error
				price, err = ParseStringToFloat(priceString)
				if err != nil {
					return names, fmt.Errorf("ошибка при парсинге цены товара: %v", err)
				}
			}

			fmt.Println("priceString", priceString)

			names = append(names, modelsI.GoodElement{Name: trimmedName, Size: findedSize, Price: price})
			firstindex = 0
		}

	}

	return names, nil
}
