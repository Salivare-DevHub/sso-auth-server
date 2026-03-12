package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// EnvLocal, EnvDev, and EnvProd define environment names.
const (
	EnvLocal = "local"
	EnvDev   = "dev"
	EnvProd  = "prod"
)

// Config holds application configuration.
type Config struct {
	Env      string        `yaml:"env" env-default:"local"`
	TokenTTL time.Duration `yaml:"token_ttl" env-required:"true"`
	HTTP     HTTPConfig    `yaml:"http"`
	Redis    RedisConfig   `yaml:"redis"`

	OAuthProviders map[string]OAuthProvider  `yaml:"-"`
	InternalApps   map[string]AppCredentials `yaml:"-"`
}

// HTTPConfig defines HTTP server settings.
type HTTPConfig struct {
	Host            string        `yaml:"host" env-required:"true"`
	Port            int           `yaml:"port"`
	Timeout         time.Duration `yaml:"timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// RedisConfig defines Redis connection settings.
type RedisConfig struct {
	Host string `yaml:"host" env-required:"true"`
	Port int    `yaml:"port" env-required:"true"`
}

// OAuthProvider describes OAuth provider settings.
type OAuthProvider struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"userinfo_url"`
	Scopes       []string `json:"scopes"`
}

// AppCredentials holds app credentials and token settings.
type AppCredentials struct {
	ClientID    string        `json:"client_id"`
	BearerToken string        `json:"bearer_token"`
	SigningKey  string        `json:"signing_key"`
	AccessTTL   time.Duration `json:"access_ttl"`
	RefreshTTL  time.Duration `json:"refresh_ttl"`
}

// UnmarshalJSON decodes AppCredentials from JSON with time.Duration handling.
func (a *AppCredentials) UnmarshalJSON(data []byte) error {
	type auxAppCredentials struct {
		ClientID    string `json:"client_id"`
		BearerToken string `json:"bearer_token"`
		SigningKey  string `json:"signing_key"`
		AccessTTL   string `json:"access_ttl"`
		RefreshTTL  string `json:"refresh_ttl"`
	}

	var aux auxAppCredentials
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	a.ClientID = aux.ClientID
	a.BearerToken = aux.BearerToken
	a.SigningKey = aux.SigningKey

	var errAccess error
	if aux.AccessTTL != "" {
		a.AccessTTL, errAccess = time.ParseDuration(aux.AccessTTL)
		if errAccess != nil {
			return fmt.Errorf("invalid access_ttl: %w", errAccess)
		}
	}

	var errRefresh error
	if aux.RefreshTTL != "" {
		a.RefreshTTL, errRefresh = time.ParseDuration(aux.RefreshTTL)
		if errRefresh != nil {
			return fmt.Errorf("invalid refresh_ttl: %w", errRefresh)
		}
	}

	return nil
}

// MustLoad loads config from the default path or panics.
func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config file path is empty")
	}

	return MustLoadByPath(path)
}

// MustLoadByPath loads config from a path or panics.
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
	if err := validateOAuthProviders(cfg.OAuthProviders); err != nil {
		panic("invalid OAUTH_PROVIDERS: " + err.Error())
	}

	internalJSON := os.Getenv("INTERNAL_APPS")
	if internalJSON == "" {
		panic("INTERNAL_APPS env variable is required")
	}

	if err := json.Unmarshal([]byte(internalJSON), &cfg.InternalApps); err != nil {
		panic("failed to parse INTERNAL_APPS: " + err.Error())
	}
	if err := validateInternalApps(cfg.InternalApps); err != nil {
		panic("invalid INTERNAL_APPS: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath resolves the config path from flags or env.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

func validateOAuthProviders(providers map[string]OAuthProvider) error {
	if len(providers) == 0 {
		return errors.New("no providers configured")
	}

	for name, p := range providers {
		if p.ClientID == "" {
			return fmt.Errorf("%s: client_id is required", name)
		}
		if p.ClientSecret == "" {
			return fmt.Errorf("%s: client_secret is required", name)
		}
		if p.RedirectURL == "" {
			return fmt.Errorf("%s: redirect_url is required", name)
		}
		if p.TokenURL == "" {
			return fmt.Errorf("%s: token_url is required", name)
		}
		if p.UserInfoURL == "" {
			return fmt.Errorf("%s: userinfo_url is required", name)
		}
		if len(p.Scopes) == 0 {
			return fmt.Errorf("%s: scopes is required", name)
		}
	}

	return nil
}

func validateInternalApps(apps map[string]AppCredentials) error {
	if len(apps) == 0 {
		return errors.New("no apps configured")
	}

	for name, a := range apps {
		if a.ClientID == "" {
			return fmt.Errorf("%s: client_id is required", name)
		}
		if a.BearerToken == "" {
			return fmt.Errorf("%s: bearer_token is required", name)
		}
		if a.SigningKey == "" {
			return fmt.Errorf("%s: signing_key is required", name)
		}
	}

	return nil
}
