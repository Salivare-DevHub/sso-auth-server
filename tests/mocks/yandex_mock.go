package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

// YandexServerMock replaces the real Yandex OAuth server.
type YandexServerMock struct {
	server *httptest.Server
	mu     sync.Mutex
	codes  map[string]YandexUserInfo // code -> user info
}

type YandexUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RealName  string `json:"real_name"`
	Email     string `json:"default_email"`
	Picture   string `json:"default_avatar_id"`
}

type YandexTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// NewYandexServerMock creates a YandexServerMock.
func NewYandexServerMock() *YandexServerMock {
	mock := &YandexServerMock{
		codes: make(map[string]YandexUserInfo),
	}

	mock.server = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/token" {
					mock.handleToken(w, r)
				} else if r.URL.Path == "/info" {
					mock.handleUserInfo(w, r)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			},
		),
	)

	return mock
}

func (m *YandexServerMock) handleToken(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	m.mu.Lock()
	_, codeExists := m.codes[code]
	m.mu.Unlock()

	if !codeExists {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_code"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(
		YandexTokenResponse{
			AccessToken: "mock_token_" + code,
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		},
	)
}

func (m *YandexServerMock) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Extract code from token (Bearer mock_token_CODE).
	var code string
	_, _ = fmt.Sscanf(token, "Bearer mock_token_%s", &code)

	m.mu.Lock()
	userInfo, exists := m.codes[code]
	m.mu.Unlock()

	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(userInfo)
}

// RegisterCode registers an OAuth code with user info.
func (m *YandexServerMock) RegisterCode(code string, userInfo YandexUserInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[code] = userInfo
}

// URL returns the mock server base URL.
func (m *YandexServerMock) URL() string {
	return m.server.URL
}

// Close shuts down the mock server.
func (m *YandexServerMock) Close() {
	m.server.Close()
}
