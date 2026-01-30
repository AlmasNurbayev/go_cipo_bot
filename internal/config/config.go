package config

import (
	"log"
	"time"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/lib/utils"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	POSTGRES_USER     string        `env:"POSTGRES_USER" json:"-"`
	POSTGRES_PASSWORD string        `env:"POSTGRES_PASSWORD" json:"-"`
	POSTGRES_DB       string        `env:"POSTGRES_DB"`
	POSTGRES_PORT     string        `env:"POSTGRES_PORT"`
	POSTGRES_TIMEOUT  time.Duration `env:"POSTGRES_TIMEOUT"`
	POSTGRES_HOST     string        `env:"POSTGRES_HOST"`
	POSTGRES_INT_PORT string        `env:"POSTGRES_INT_PORT"`

	SECRET_KEY     string `env:"SECRET_KEY"  json:"-"` // для шифрования в БД
	SECRET_BYTE    []byte `json:"-"`                   // для шифрования в БД
	GOOGLE_API_KEY string `env:"GOOGLE_API_KEY" json:"-"`

	BOT_TOKEN   string        `env:"BOT_TOKEN"  json:"-"`
	BOT_TIMEOUT time.Duration `env:"BOT_TIMEOUT"`
	HTTP_PORT   string        `env:"HTTP_PORT"`

	KOFD_PASSAUTH_URL   string `env:"KOFD_PASSAUTH_URL"`
	KOFD_OPERATIONS_URL string `env:"KOFD_OPERATIONS_URL"`

	KAFKA_PORT         string `env:"KAFKA_PORT"`
	KAFKA_SERVICE_NAME string `env:"KAFKA_SERVICE_NAME"`

	NATS_NAME            string `env:"NATS_NAME"`
	NATS_PORT            string `env:"NATS_PORT"`
	NATS_MONITORING_PORT string `env:"NATS_MONITORING_PORT"`
	NATS_STREAM_NAME     string `env:"NATS_STREAM_NAME"`
	NATS_ENABLE          bool   `env:"NATS_ENABLE" envDefault:"true"`

	CIPO_PRODUCTS_URL string `env:"CIPO_PRODUCTS_URL"` // в бэкенд CIPO - карточка товара
	CIPO_IMAGES_URL   string `env:"CIPO_IMAGES_URL"`   // в бэкенд CIPO - статика
	CIPO_QNT_URL      string `env:"CIPO_QNT_URL"`      // в бэкенд CIPO - остатки и цены склада

	LOG_ERROR_PATH string `env:"LOG_ERROR_PATH"`

	ENV string `env:"ENV"`
}

func Mustload(path string) *Config {
	cfg := &Config{}

	if path != "" {
		err := godotenv.Load(path)
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal("not load config: ", err)
	}

	if cfg.SECRET_KEY != "" {
		cfg.SECRET_BYTE = utils.DeriveKeyFromSecret(cfg.SECRET_KEY)
	}

	if cfg.LOG_ERROR_PATH == "" {
		cfg.LOG_ERROR_PATH = "_volume_assets/error.log"
	}

	return cfg
}
