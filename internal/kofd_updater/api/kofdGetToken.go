package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/config"
)

type KofdAuthenticateRequest struct {
	Credentials     Credentials `json:"credentials"`
	OrganizationXin string      `json:"organizationXin"`
}

type Credentials struct {
	Iin      string `json:"iin"`
	Password string `json:"password"`
}

type KofdAuthenticateResponse struct {
	Data  KofdAuthenticateResponseData `json:"data"`
	Error any                          `json:"error"`
}

type KofdAuthenticateResponseData struct {
	Jwt                   string `json:"jwt"`
	AccessTokenExpiration string `json:"accessTokenExpiration"`
	RefreshToken          string `json:"refreshToken"`
}

func KofdGetToken(sendBody KofdAuthenticateRequest, cfg *config.Config,
	log *slog.Logger) (KofdAuthenticateResponse, error) {

	var response = KofdAuthenticateResponse{}

	op := "api.KofdGetToken"
	log = log.With(slog.String("op", op))

	base, err := url.Parse(cfg.KOFD_PASSAUTH_URL)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}
	jsonBody, err := json.Marshal(sendBody)
	if err != nil {
		log.Error("Error marshaling request body:", slog.String("err", err.Error()))
		return response, err
	}
	req, err := http.NewRequest("POST", base.String(), io.NopCloser(bytes.NewReader(jsonBody)))
	client := &http.Client{}

	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}
	req.Header.Set("Content-Type", "application/json")
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
		return response, err
	}

	return response, nil
}
