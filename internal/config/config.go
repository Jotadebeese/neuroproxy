package config

import (
	"os"
	"strings"
)

// Config holds all runtime configuration for the NeuroProxy gateway.
type Config struct {
	Port         string
	GeminiAPIKey string
	GCPProject   string
	RedisURL     string
	StaticDir    string
	Environment  string
}

// Load reads configuration from environment variables and an optional .env file.
func Load() *Config {
	loadDotEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	gcpProject := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if gcpProject == "" {
		gcpProject = os.Getenv("GCP_PROJECT")
	}
	if gcpProject == "" {
		gcpProject = "neuroproxy-prod"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "frontend/out"
	}

	return &Config{
		Port:         port,
		GeminiAPIKey: strings.TrimSpace(os.Getenv("GEMINI_API_KEY")),
		GCPProject:   gcpProject,
		RedisURL:     strings.TrimSpace(os.Getenv("REDIS_URL")),
		StaticDir:    staticDir,
		Environment:  env,
	}
}

func loadDotEnv(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}
