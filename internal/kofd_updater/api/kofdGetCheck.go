package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
)

type KofdCheckRowResponse struct {
	Text  string `json:"text"`
	Style int    `json:"style"`
}

type KofdCheckResponseData struct {
	Data  []KofdCheckRowResponse `json:"data"`
	Error any                    `json:"error"`
}

func KofdGetCheck(cfg *config.Config,
	log1 *slog.Logger, knumber string, token string,
	id string) (KofdCheckResponseData, error) {

	var response = KofdCheckResponseData{}

	op := "api.KofdGetCheck"
	log := log1.With(slog.String("op", op))

	base, err := url.Parse(cfg.KOFD_OPERATIONS_URL + "/operation")
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}
	query := url.Values{}
	query.Add("cashboxId", knumber)
	query.Add("operationId", id)
	base.RawQuery = query.Encode()

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

	return response, nil

}
