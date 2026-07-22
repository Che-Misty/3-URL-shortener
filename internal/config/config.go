package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	DatabaseURL string `yaml:"database_url" env:"DATABASE_URL"`
	AliasLength int    `yaml:"alias_length" env:"ALIAS_LEN"`
	HTTPServer  `yaml:"http_server"`
	Auth        `yaml:"auth"`
}

type Auth struct {
	User     string `yaml:"user" env:"AUTH_USER"`
	Password string `yaml:"password" env:"AUTH_PASSWORD"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"0.0.0.0:8082"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("could not load .env: %v", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DatabaseURL = dbURL
		log.Println("Using DATABASE_URL from environment")
	}

	if user := os.Getenv("AUTH_USER"); user != "" {
		cfg.Auth.User = user
		log.Println("Using AUTH_USER from environment")
	}

	if password := os.Getenv("AUTH_PASSWORD"); password != "" {
		cfg.Auth.Password = password
		log.Println("Using AUTH_PASSWORD from environment")
	}

	if alias_lengthStr := os.Getenv("ALIAS_LEN"); alias_lengthStr != "" {
		alias_length, err := strconv.Atoi(alias_lengthStr)
		if err != nil {
			log.Printf("could not parse ALIAS_LEN: %v", err)
		}
		cfg.AliasLength = alias_length
		log.Println("Using ALIAS_LEN from environment")
	}

	return &cfg
}
