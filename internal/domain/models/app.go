package models

import "time"

type App struct {
	ID         int64         `yaml:"id"`
	Secret     string        `yaml:"secret"`
	Name       string        `yaml:"name"`
	AccessTTL  time.Duration `yaml:"access_ttl"`
	RefreshTTL time.Duration `yaml:"refresh_ttl"`
}
