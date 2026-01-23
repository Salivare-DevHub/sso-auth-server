package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env           string        `yaml:"env" env-default:"local"`
	TokenTTL      time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC          GRPCConfig    `yaml:"grpc"`
	AuthProviders AuthConfig    `yaml:"auth_providers" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type AuthConfig struct {
	Google GoogleConfig `yaml:"google"`
	Yandex YandexConfig `yaml:"yandex"`
}

type GoogleConfig struct {
	ClientID     string `yaml:"client_id" env-required:"true"`
	ClientSecret string `yaml:"client_secret" env-required:"true"`
	RedirectURL  string `yaml:"redirect_url" env-required:"true"`
}

type YandexConfig struct {
	ClientID     string `yaml:"client_id" env-required:"true"`
	ClientSecret string `yaml:"client_secret" env-required:"true"`
	RedirectURL  string `yaml:"redirect_url" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchConfigPatch()

	if path == "" {
		panic("config file path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPatch() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATCH")
	}

	return res
}
