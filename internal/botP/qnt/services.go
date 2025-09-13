package qnt

import (
	"log/slog"
	"sort"

	botP "github.com/AlmasNurbayev/go_cipo_bot/internal/botP/api"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
)

func qntNowService(log1 *slog.Logger, cfg *config.Config) (string, error) {

	op := "summary.getAnalytics"
	log := log1.With(slog.String("op", op))

	qntData, err := botP.CipoProductsOnlyQnt(cfg, log, "")
	if err != nil {
		log.Error("error on get qntData from Cipo backend: ", slog.String("err", err.Error()))
	}

	groupData := transformQntData(qntData.Products)

	text := "Остатки на сегодня:\n"
	text += "\n"
	text += "<b>Магазины:</b> \n"
	for _, v := range groupData.Stores {
		text += " • " + v.Name + ": " + utils.FormatNumber(float64(v.Qnt)) + " шт. \n" +
			" на сумму " + utils.FormatNumber(float64(v.Sum)) + " \n"
	}
	text += "\n"
	text += "<b>Товары:</b> \n"
	text += " Виды номенклатуры: \n"
	for _, v := range groupData.Goods.NomVids {
		text += "  • " + v.Name + ": " + utils.FormatNumber(float64(v.Qnt)) + " шт. " +
			" на сумму " + utils.FormatNumber(float64(v.Sum)) + " \n"

		for _, pg := range v.ProductGroups {
			text += "<b>    ◦ " + pg.Name + ": " + utils.FormatNumber(float64(pg.Qnt)) + " шт. " +
				" на сумму " + utils.FormatNumber(float64(pg.Sum)) + " </b>\n"
			sizes := pg.SizeGroups
			sort.Slice(sizes, func(i, j int) bool {
				return sizes[i].Name < sizes[j].Name
			})
			for _, sg := range sizes {
				text += "      - " + sg.Name + ": " + utils.FormatNumber(float64(sg.Qnt)) + " шт. " +
					" на сумму " + utils.FormatNumber(float64(sg.Sum)) + " \n"
			}
			text += "\n"
		}
		text += "\n"
	}

	return text, nil

}
