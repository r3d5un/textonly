package config_test

import (
	"fmt"
	"testing"

	"textonly.islandwind.me/cmd/web/config"
)

func TestNewConfig(t *testing.T) {
	config, err := config.New()
	if err != nil {
		t.Errorf("could not create new config '%v'", err)
	}

	fmt.Println(config.Database.DSN)

	if config.App.ENV != "development" {
		t.Errorf("expected 'development', got '%v'", config.App.ENV)
	}
	if config.App.URL != ":4000" {
		t.Errorf("expected ':4000', got '%v'", config.App.URL)
	}
	if config.App.Password != "password" {
		t.Errorf("expected 'password', got '%v'", config.App.Password)
	}
	if config.App.User != "admin" {
		t.Errorf("expected 'admin', got '%v'", config.App.User)
	}
	if config.Database.DSN != "postgresql://postgres:postgres@database:5432?database=blog" {
		t.Errorf(
			"expected 'postgresql://postgres:postgres@database:5432?database=blog', got '%v'",
			config.Database.DSN,
		)
	}
}
