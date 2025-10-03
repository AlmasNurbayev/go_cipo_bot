package finance

import (
	"log/slog"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	modelsI "github.com/AlmasNurbayev/go_cipo_bot/internal/models"
	"github.com/kr/pretty"
)

func financeOPIUService(mode string, settings []modelsI.SettingsEntity,
	log1 *slog.Logger) (string, error) {

	op := "finance.financeOPIUService"
	log := log1.With(slog.String("op", op), slog.String("mode", mode))
	var result string

	// Получаем границы текущего дня в локальном времени
	start, end, err := utils.GetPeriodByMode(mode)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}

	books := config.GetSettingsGSheetsSources("FINANCE_GHEETS_SOURCES", settings)
	opiu_special_items := config.GetSettingsString("FINANCE_OPIU_SPECIAL_ITEMS", settings)
	finance_opiu_cost_items := config.GetSettingsString("FINANCE_OPIU_COST_ITEMS", settings)
	finance_opiu_special_items := config.GetSettingsString("FINANCE_OPIU_SPECIAL_ITEMS", settings)

	pretty.Log(books)
	pretty.Log(opiu_special_items)
	pretty.Log(finance_opiu_cost_items)
	pretty.Log(finance_opiu_special_items)
	pretty.Log(start)
	pretty.Log(end)

	return result, nil
}
