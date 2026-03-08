package config

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

type Config struct {
	Env      string        `yaml:"env" env-default:"local"`
	TokenTTL time.Duration `yaml:"token_ttl" env-required:"true"`
	HTTP     HTTPConfig    `yaml:"http"`
	Redis    RedisConfig   `yaml:"redis"`

	OAuthProviders map[string]OAuthProvider  `yaml:"-"`
	InternalApps   map[string]AppCredentials `yaml:"-"`
}

type HTTPConfig struct {
	Host            string        `yaml:"host" env-required:"true"`
	Port            int           `yaml:"port"`
	Timeout         time.Duration `yaml:"timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type RedisConfig struct {
	Host string `yaml:"host" env-required:"true"`
	Port int    `yaml:"port" env-required:"true"`
}

type OAuthProvider struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

type AppCredentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func MustLoad() *Config {
	path := fetchConfigPath()

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

	oauthJSON := os.Getenv("OAUTH_PROVIDERS")
	if oauthJSON == "" {
		panic("OAUTH_PROVIDERS env variable is required")
	}

	if err := json.Unmarshal([]byte(oauthJSON), &cfg.OAuthProviders); err != nil {
		panic("failed to parse OAUTH_PROVIDERS: " + err.Error())
	}

	internalJSON := os.Getenv("INTERNAL_APPS")
	if internalJSON == "" {
		panic("INTERNAL_APPS env variable is required")
	}

	if err := json.Unmarshal([]byte(internalJSON), &cfg.InternalApps); err != nil {
		panic("failed to parse INTERNAL_APPS: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
