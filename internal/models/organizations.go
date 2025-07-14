package models

import "github.com/guregu/null/v5"

type OrganizationEntity struct {
	Id   int64       `json:"id" db:"id"`
	Bin  string      `json:"bin" db:"bin"`
	Name string      `json:"name" db:"name"`
	Hash null.String `json:"hash" db:"hash"`
}
