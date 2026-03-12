package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

// GoogleServerMock replaces the real Google OAuth server.
type GoogleServerMock struct {
	server *httptest.Server
	mu     sync.Mutex
	codes  map[string]GoogleUserInfo // code -> user info
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type GoogleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewGoogleServerMock creates a GoogleServerMock.
func NewGoogleServerMock() *GoogleServerMock {
	mock := &GoogleServerMock{
		codes: make(map[string]GoogleUserInfo),
	}

	mock.server = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/token" {
					mock.handleToken(w, r)
				} else if r.URL.Path == "/v1/userinfo" {
					mock.handleUserInfo(w, r)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			},
		),
	)

	return mock
}

func (m *GoogleServerMock) handleToken(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	fmt.Printf("[mock] token request: code=%q form=%v\n", code, r.Form)
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
		GoogleTokenResponse{
			AccessToken: "mock_token_" + code,
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		},
	)
}

func (m *GoogleServerMock) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Extract code from token (mock_token_CODE).
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
func (m *GoogleServerMock) RegisterCode(code string, userInfo GoogleUserInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[code] = userInfo
}

// URL returns the mock server base URL.
func (m *GoogleServerMock) URL() string {
	return m.server.URL
}

// Close shuts down the mock server.
func (m *GoogleServerMock) Close() {
	m.server.Close()
}
