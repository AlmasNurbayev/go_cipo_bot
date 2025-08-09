package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/guregu/null/v5"
)

type TransactionEntity struct {
	Id                  int64          `json:"id" db:"id"`
	Ofd_id              string         `json:"ofd_id" db:"ofd_id"`
	Ofd_name            null.String    `json:"ofd_name" db:"ofd_name"`
	Onlinefiscalnumber  null.Int64     `json:"onlinefiscalnumber" db:"onlinefiscalnumber"`
	Offlinefiscalnumber null.Int64     `json:"offlinefiscalnumber" db:"offlinefiscalnumber"`
	Systemdate          null.Time      `json:"systemdate" db:"systemdate"`
	Operationdate       null.Time      `json:"operationdate" db:"operationdate"`
	Type_operation      int            `json:"type_operation" db:"type_operation"`
	Subtype             null.Int       `json:"subtype" db:"subtype"`
	Sum_operation       null.Float     `json:"sum_operation" db:"sum_operation"`
	Availablesum        null.Float     `json:"availablesum" db:"availablesum"`
	Paymenttypes        *[]int         `json:"paymenttypes" db:"paymenttypes"`
	Shift               null.Int       `json:"shift" db:"shift"`
	Created_at          time.Time      `json:"created_at" db:"created_at"`
	Organization_id     int64          `json:"organization_id" db:"organization_id"`
	Kassa_id            int64          `json:"kassa_id" db:"kassa_id"`
	Kassa_name          null.String    `json:"kassa_name" db:"kassa_name"`
	Knumber             null.String    `json:"knumber" db:"knumber"`
	Cheque              null.String    `json:"cheque" db:"cheque"`
	Images              *[]string      `json:"images" db:"images"`
	Names               *[]string      `json:"names" db:"names"`
	ChequeJSON          ChequeJSONList `json:"cheque_json" db:"cheque_json"`
}

type ChequeJSONElement struct {
	Name          string      `json:"name" db:"name"`
	Season        null.String `json:"season" db:"season"`
	VidModeli     null.String `json:"vid_modeli" db:"vid_modeli"`
	MainImageURL  null.String `json:"main_image_url" db:"main_image_url"`
	Size          null.String `json:"size" db:"size"`
	NominalPrice  float64     `json:"nominal_price" db:"nominal_price"`
	DiscountPrice float64     `json:"discount_price" db:"discount_price"`
	Sum           float64     `json:"sum" db:"sum"`
	Qnt           int         `json:"qnt" db:"qnt"`
}

type ChequeJSONList []ChequeJSONElement

func (c *ChequeJSONList) Scan(src interface{}) error {
	if src == nil {
		*c = nil
		return nil
	}

	var data []byte

	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("ChequeJSONList: unsupported type %T", src)
	}

	return json.Unmarshal(data, c)
}
