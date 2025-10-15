package botP

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

func GsheetsData(googleApiKey string, bookID string, sheet string, rangeName string, log1 *slog.Logger) ([]models.GSheetsEntityV1, error) {
	var result = []models.GSheetsEntityV1{}
	var response = models.GSheetsResponseV1{}

	op := "api.GsheetsData"
	log := log1.With(slog.String("op", op))

	base, err := url.Parse("https://sheets.googleapis.com/v4/spreadsheets/" + bookID + "/values/" + sheet + "!" + rangeName)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return result, err
	}
	query := url.Values{}
	query.Add("key", googleApiKey)
	query.Add("valueRenderOption", "FORMATTED_VALUE")
	base.RawQuery = query.Encode()
	//log.Debug("request", slog.String("url", base.String()))

	req, err := http.NewRequest("GET", base.String(), nil)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return result, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return result, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("Error closing response body:", slog.String("err", err.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Error("Api error:", slog.String("err", resp.Status))
		return result, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return result, err
	}

	if err := json.Unmarshal(resBody, &response); err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return result, err
	}

	result, err = convertResponseToEntitiesV1(response)
	if err != nil {
		log.Error("error: ", slog.String("err", err.Error()))
		return result, err
	}

	return result, nil
}

func convertResponseToEntitiesV1(response models.GSheetsResponseV1) ([]models.GSheetsEntityV1, error) {
	entities := []models.GSheetsEntityV1{}
	if len(response.Values) < 1 {
		return entities, nil
	}
	for _, row := range response.Values[0:] {
		entity := models.GSheetsEntityV1{}
		for i, cell := range row {
			switch i {
			case 0:
				if cell == "" {
					continue
				}
				s, err := utils.ParseStringToInt(cell)
				if err != nil {
					return entities, fmt.Errorf("%w Id for Id=%d", err, i)
				}
				entity.Id = s
			case 1:
				entity.Description = cell
			case 2:
				entity.Category = cell
			case 3:
				entity.Division = cell
			case 4:
				if cell == "" {
					entity.Period = time.Time{}
				} else {
					t, err := time.Parse("02.01.2006", cell)
					if err != nil {
						return entities, fmt.Errorf("%w Id=%d", err, entity.Id)
					}
					entity.Period = t
				}
			case 5:
				if cell == "" {
					entity.Date = time.Time{}
				} else {
					t, err := time.Parse("02.01.2006", cell)
					if err != nil {
						return entities, fmt.Errorf("%w Date for Id=%d", err, entity.Id)
					}
					entity.Date = t
				}
			case 6:
				if cell == "" {
					entity.Sum = 0
				} else {
					s, err := utils.ParseStringToFloat(cell)
					if err != nil {
						return entities, fmt.Errorf("%w Sum for Id=%d", err, entity.Id)
					}
					entity.Sum = s
				}
			case 7:
				entity.Currency = cell
			case 8:
				if cell == "" {
					entity.SumUSD = 0
				} else {
					s, err := utils.ParseStringToFloat(cell)
					if err != nil {
						return entities, fmt.Errorf("%w SumUSD for Id=%d", err, entity.Id)
					}
					entity.SumUSD = s
				}
			case 9:
				entity.Account = cell
			case 10:
				entity.Organization = cell
				// Add more cases if there are more columns
			}
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
