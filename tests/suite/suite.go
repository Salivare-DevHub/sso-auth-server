package suite

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/salivare-io/slogx"

	authconnect "github.com/salivare-io/protos-sso/gen/go/sso/service/auth/v1/authservicev1connect"
	"github.com/salivare-io/sso-auth-server/internal/app"
	"github.com/salivare-io/sso-auth-server/internal/config"
)

// Suite holds shared test fixtures.
type Suite struct {
	*testing.T
	Cfg        *config.Config
	HTTPClient *http.Client
	AuthClient authconnect.AuthServiceClient
	App        *app.App
}

// New creates the test context, suite, and starts the test server.
func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()

	configPath := resolveConfigPath()
	cfg := config.MustLoadByPath(configPath)

	// Find a free port for the test server.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	freePort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Override the port in config.
	cfg.HTTP.Port = freePort

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	httpClient := &http.Client{
		Timeout: 25 * time.Second,
	}

	// Create the server logger.
	log := slogx.New(
		slogx.WithLevel(slogx.LevelTrace),
		slogx.WithContextKeys("trace_id", "request_id"),
		slogx.WithRemoval(slogx.NewRemovalSet().Add("bearer_token")),
	)

	// Initialize the application.
	application, err := app.New(log, cfg)
	if err != nil {
		t.Fatalf("failed to init application: %v", err)
	}

	// Start the server in a separate goroutine.
	go func() {
		if err := application.RPCSrv.Run(); err != nil {
			log.Error("server run error", slog.String("err", err.Error()))
		}
	}()

	// Verify the server started.
	addr := net.JoinHostPort(cfg.HTTP.Host, strconv.Itoa(cfg.HTTP.Port))
	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		if i == 9 {
			t.Fatalf("failed to connect to server at %s", addr)
		}
		time.Sleep(100 * time.Millisecond)
	}

	suite := &Suite{
		T:          t,
		Cfg:        cfg,
		HTTPClient: httpClient,
		AuthClient: authconnect.NewAuthServiceClient(httpClient, "http://"+addr),
		App:        application,
	}

	// Register cleanup to stop the server.
	t.Cleanup(func() {
		suite.App.RPCSrv.Stop()
	})

	return ctx, suite
}

// resolveConfigPath returns the config path.
// Priority order:
// 1. CONFIG_PATH (env var)
// 2. Lookup relative to suite.go.
// 3. Lookup in standard locations.
func resolveConfigPath() string {
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		return envPath
	}

	// Get the current file path.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "./configs/testlocal.yaml"
	}

	// Find the project root (folder with testlocal.yaml).
	projectRoot := filepath.Dir(filepath.Dir(filename))
	configPath := filepath.Join(projectRoot, "..", "configs", "testlocal.yaml")

	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Fallback: try standard paths.
	standardPaths := []string{
		"./configs/testlocal.yaml",
		"../configs/testlocal.yaml",
		"../../configs/testlocal.yaml",
	}

	for _, path := range standardPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// If nothing found, return default path (config.MustLoadByPath will panic).
	return "./configs/testlocal.yaml"
}

// URL builds a full URL to the test service.
func (s *Suite) URL(path string) string {
	addr := net.JoinHostPort(s.Cfg.HTTP.Host, strconv.Itoa(s.Cfg.HTTP.Port))
	return "http://" + addr + path
}

// GetAppCredentials returns app credentials by name.
// Returns (credentials, found).
func (s *Suite) GetAppCredentials(appName string) (config.AppCredentials, bool) {
	if s == nil || s.Cfg == nil || s.Cfg.InternalApps == nil {
		return config.AppCredentials{}, false
	}

	creds, ok := s.Cfg.InternalApps[appName]
	return creds, ok
}
