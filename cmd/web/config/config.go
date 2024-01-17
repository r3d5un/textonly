package config

import (
	_ "embed"

	"github.com/spf13/viper"
)

type Config struct {
	Database *DatabaseConfig `json:"database"`
	App      *AppConfig      `json:"app"`
}

type DatabaseConfig struct {
	DSN     string `json:"-"`
	Timeout int    `json:"timeout"`
}

type AppConfig struct {
	URL      string `json:"url"`
	ENV      string `json:"env"`
	User     string `json:"user"`
	Password string `json:"-"`
	Realm    string `json:"realm"`
}

func New() (*Config, error) {
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(false)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/app/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.BindEnv("database.dsn", "TEXTONLY_DSN")
	if err != nil {
		return nil, err
	}

	err = viper.BindEnv("app.url", "TEXTONLY_URL")
	if err != nil {
		return nil, err
	}
	err = viper.BindEnv("app.env", "TEXTONLY_ENV")
	if err != nil {
		return nil, err
	}

	err = viper.BindEnv("app.user", "TEXTONLY_USER")
	if err != nil {
		return nil, err
	}
	err = viper.BindEnv("app.password", "TEXTONLY_PASSWORD")
	if err != nil {
		return nil, err
	}
	err = viper.BindEnv("app.password", "TEXTONLY_REALM")
	if err != nil {
		return nil, err
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
