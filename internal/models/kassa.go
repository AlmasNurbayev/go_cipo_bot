package models

import "github.com/guregu/null/v5"

type KassaEntity struct {
	Id              int64       `json:"id" db:"id"`
	Snumber         string      `json:"snumber" db:"snumber"`
	Znumber         string      `json:"znumber" db:"znumber"`
	Knumber         null.String `json:"knumber" db:"knumber"`
	Name_kassa      string      `json:"name_kassa" db:"name_kassa"`
	Organization_id int64       `json:"organization_id" db:"organization_id"`
	Is_active       bool        `json:"is_active" db:"is_active"`
}
