package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
	"github.com/guregu/null/v5"
)

type KofdOperationsResponse struct {
	Data       []KofdOperationsResponseData `json:"data"`
	GroupCount int                          `json:"groupCount"`
	TotalCount int                          `json:"totalCount"`
	Summary    any                          `json:"summary"`
}

type KofdOperationsResponseData struct {
	Id                  string     `json:"id"`
	AvailableSum        float64    `json:"availableSum"`
	OfflineFiscalNumber null.Int64 `json:"offlineFiscalNumber"`
	OnlineFiscalNumber  null.Int64 `json:"onlineFiscalNumber"`
	OperationDate       time.Time  `json:"operationDate"`
	PaymentTypes        []int      `json:"paymentTypes"`
	Shift               int        `json:"shift"`
	SubType             int        `json:"subType"`
	Sum_operation       float64    `json:"sum"`
	SystemDate          time.Time  `json:"systemDate"`
	Type_operation      int        `json:"type"`
}

func KofdGetOperations(cfg *config.Config,
	log *slog.Logger, knumber string, token string,
	firstDate string, lastDate string) (KofdOperationsResponse, error) {

	var response = KofdOperationsResponse{}

	op := "kofd.KofdGetOperations"
	log = log.With(slog.String("op", op))

	base, err := url.Parse(cfg.KOFD_OPERATIONS_URL)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}
	query := url.Values{}
	query.Add("cashboxId", knumber)
	query.Add("toDate", lastDate+"T23:59:59")
	query.Add("fromDate", firstDate)
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

	return response, nil
}
