package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Env           string `env:"APP_ENV" envDefault:"local"`
	HTTPAddr      string `env:"HTTP_ADDR" envDefault:":8080"`
	DatabaseURL   string `env:"DATABASE_URL,required"`
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	DeepSeekApiKey  string `env:"DEEPSEEK_API_KEY"`
	DeepSeekBaseURL string `env:"DEEPSEEK_BASE_URL" envDefault:"https://api.deepseek.com"`
	DeepSeekModel   string `env:"DEEPSEEK_MODEL" envDefault:"deepseek-chat"`
}

func Load() (Config, error) {
	_ = godotenv.Load()

	var cfg Config
	err := env.Parse(&cfg)
	return cfg, err
}