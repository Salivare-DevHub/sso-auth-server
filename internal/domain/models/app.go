package models

import "time"

type App struct {
	ID         string
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}
