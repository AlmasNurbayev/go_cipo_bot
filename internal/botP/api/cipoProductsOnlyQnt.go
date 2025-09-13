package botP

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

func CipoProductsOnlyQnt(cfg *config.Config,
	log1 *slog.Logger, token string) (models.ProductsOnlyQntResponse, error) {

	var response = models.ProductsOnlyQntResponse{}

	op := "api.CipoProductsOnlyQnt"
	log := log1.With(slog.String("op", op))

	base, err := url.Parse(cfg.CIPO_QNT_URL)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}
	// query := url.Values{}
	// query.Add("name_1c", name_1c)
	// base.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", base.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("Error closing response body:", slog.String("err", err.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Error("Api error:", slog.String("err", resp.Status))
		return response, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}

	if err := json.Unmarshal(resBody, &response); err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}

	//log.Info("Api response", slog.String("response", string(resBody)))

	//pp.Println(response)

	return response, nil

}
