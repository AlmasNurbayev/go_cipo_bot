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

type ProductByIdResponse struct {
	Id                int         `json:"id"`
	Id_1c             string      `json:"id_1c"`
	Name_1c           string      `json:"name_1c"`
	Name              string      `json:"name"`
	Base_ed           string      `json:"base_ed"`
	Artikul           string      `json:"artikul"`
	Material_up       null.String `json:"material_up"`
	Material_inside   null.String `json:"material_inside"`
	Material_podoshva null.String `json:"material_podoshva"`
	Main_color        null.String `json:"main_color"`
	Description       null.String `json:"description"`
	Sex               null.Int16  `json:"sex"`
	Product_folder    null.String `json:"product_folder"`
	Public_web        bool        `json:"public_web"`
	Product_group_id  int64       `json:"product_group_id"`
	Product_vid_id    int64       `json:"product_vid_id"`
	Vid_modeli_id     null.Int64  `json:"vid_modeli_id"`
	Registrator_id    int64       `json:"registrator_id"`

	Changed_date null.Time `json:"changed_date" db:"changed_date"`
	Create_date  time.Time `json:"create_date" db:"create_date"`

	Product_group            idName1c                `json:"product_group"`
	Vid_modeli               idName1c                `json:"vid_modeli"`
	Image_registry           []ImageRegistryResponse `json:"image_registry"`
	Qnt_price_registry       []qntPriceRegistry      `json:"qnt_price_registry"`
	Qnt_price_registry_group []qntPriceRegistryGroup `json:"qnt_price_registry_group"`
}

type idName1c struct {
	Id      int    `json:"id"`
	Name_1c string `json:"name_1c"`
}

type ImageRegistryResponse struct {
	Id                 int64       `json:"id"`
	Resolution         null.String `json:"resolution"`
	Full_name          string      `json:"full_name"`
	Name               string      `json:"name"`
	Path               string      `json:"path"`
	Size               int         `json:"size"`
	Operation_date     string      `json:"operation_date"`
	Main               bool        `json:"main"`
	Main_change_date   string      `json:"main_change_date"`
	Active             bool        `json:"active"`
	Active_change_date string      `json:"active_change_date"`
	Product_id         int64       `json:"product_id"`
	Registrator_id     int64       `json:"registrator_id"`
	Create_date        string      `json:"create_date"`
	Changed_date       string      `json:"changed_date"`
}

type qntPriceRegistry struct {
	Size_id      int64   `json:"size_id"`
	Size_name_1c string  `json:"size_name_1c"`
	Qnt          float32 `json:"qnt"`
	Sum          float32 `json:"sum"`
	Store_id     int64   `json:"store_id"`
}

type qntPriceRegistryGroup struct {
	Size_id      int64   `json:"size_id"`
	Size_name_1c string  `json:"size_name_1c"`
	Qnt          float32 `json:"qnt"`
	Sum          float32 `json:"sum"`
	Store_id     []int64 `json:"store_id"`
}

func CipoGetProduct(cfg *config.Config,
	log1 *slog.Logger, name_1c string, token string) (ProductByIdResponse, error) {

	var response = ProductByIdResponse{}

	op := "api.CipoGetProduct"
	log := log1.With(slog.String("op", op))

	base, err := url.Parse(cfg.CIPO_PRODUCTS_URL)
	if err != nil {
		log.Error("Api error:", slog.String("err", err.Error()))
		return response, err
	}
	query := url.Values{}
	query.Add("name_1c", name_1c)
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
