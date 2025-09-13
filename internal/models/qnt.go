package models

import (
	"time"

	"github.com/guregu/null/v5"
)

type ProductsOnlyQntResponse struct {
	Products []ProductOnlyQnt `json:"products"`
	Count    int              `json:"count"`
}

type ProductOnlyQnt struct {
	Product_id          int64       `json:"product_id"`
	Product_name        string      `json:"product_name"`
	Vid_modeli_id       null.Int64  `json:"vid_modeli_id"`
	Vid_modeli_name     string      `json:"vid_modeli_name"`
	Product_group_id    int64       `json:"product_group_id"`
	Product_group_name  string      `json:"product_group_name"`
	Nom_vid             null.String `json:"nom_vid"`
	Store_id            int64       `json:"store_id"`
	Store_name          string      `json:"store_name"`
	Size_id             int64       `json:"size_id"`
	Size_name           string      `json:"size_name"`
	Base_ed             string      `json:"base_ed"`
	Product_create_date time.Time   `json:"product_create_date"`
	Sum                 float32     `json:"sum"`
	Sum_zakup           float32     `json:"sum_zakup"`
	Qnt                 int64       `json:"qnt"`
}
