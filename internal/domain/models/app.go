package models

import "time"

// App represents an application with token settings.
type App struct {
	ID          string        `yaml:"id"`
	BearerToken string        `yaml:"bearer_token"`
	SigningKey  string        `yaml:"signing_key"`
	Name        string        `yaml:"name"`
	AccessTTL   time.Duration `yaml:"access_ttl"`
	RefreshTTL  time.Duration `yaml:"refresh_ttl"`
}
