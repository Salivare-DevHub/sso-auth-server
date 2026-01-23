package appconfig

import (
	"context"
	"errors"
	"gopkg.in/yaml.v3"
	"os"

	"github.com/Salivare-DevHub/sso-auth-server/internal/domain/models"
)

type AppConfig struct {
	apps map[int64]models.App
}

func NewYAML(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var list []models.App
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	m := make(map[int64]models.App)
	for _, a := range list {
		m[a.ID] = a
	}

	return &AppConfig{apps: m}, nil
}

func (r *AppConfig) GetByID(ctx context.Context, id int64) (*models.App, error) {
	a, ok := r.apps[id]
	if !ok {
		return nil, errors.New("app not found")
	}
	return &a, nil
}
