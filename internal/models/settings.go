package models

import (
	"time"
)

type SettingsEntity struct {
	Key       string    `json:"key" db:"key"`
	Value     []any     `json:"value" db:"value"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Books struct {
	Book  string `json:"book"`
	Sheet string `json:"sheet"`
	Range string `json:"range"`
}
