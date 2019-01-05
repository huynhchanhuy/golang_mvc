package config

import (
	"os"
)

type Config struct {
	DB *DBConfig
}

type DBConfig struct {
	Dialect  string
	Username string
	Password string
	Hostname string
	Name     string
	Charset  string
}

func GetConfig() *Config {
	return &Config{
		DB: &DBConfig{
			Dialect:  os.Getenv("DIALECT"),
			Username: os.Getenv("DB_USERNAME"),
			Password: os.Getenv("DB_PASSWORD"),
			Hostname: os.Getenv("DB_HOST"),
			Name:     os.Getenv("DB_NAME"),
			Charset:  os.Getenv("CHARSET"),
		},
	}
}
