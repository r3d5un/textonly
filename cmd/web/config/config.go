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
	DSN string `json:"dsn"`
}

type AppConfig struct {
	URL      string `json:"url"`
	ENV      string `json:"env"`
	User     string `json:"user"`
	Password string `json:"password"`
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

	err = viper.BindEnv("database.dsn", "DSN")
	if err != nil {
		return nil, err
	}

	err = viper.BindEnv("app.url", "URL")
	if err != nil {
		return nil, err
	}
	err = viper.BindEnv("app.env", "ENV")
	if err != nil {
		return nil, err
	}

	err = viper.BindEnv("app.user", "USER")
	if err != nil {
		return nil, err
	}
	err = viper.BindEnv("app.password", "PASSWORD")
	if err != nil {
		return nil, err
	}
	err = viper.BindEnv("app.password", "REALM")
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
