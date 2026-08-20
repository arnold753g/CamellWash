package config

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv             string
	LogLevel           slog.Level
	ServerHost         string
	ServerPort         string
	DatabaseDSN        string
	AutoMigrate        bool
	AllowedOrigins     []string
	TrustProxy         bool
	SessionSecret      string
	SessionTTL         time.Duration
	SessionIdleTimeout time.Duration
	CookieSecure       bool
	CookieSameSite     http.SameSite
	PasswordBcryptCost int
	LoginMaxAttempts   int
	LoginLockout       time.Duration
	AuthRatePerMinute  int
	APIRatePerMinute   int
}

func Load() (Config, error) {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	appEnv := value("APP_ENV", "development")
	if appEnv != "development" && appEnv != "test" && appEnv != "production" {
		return Config{}, fmt.Errorf("APP_ENV must be development, test, or production")
	}

	dbPassword, err := required("DB_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	sessionSecret, err := required("SESSION_SECRET")
	if err != nil {
		return Config{}, err
	}
	if len(sessionSecret) < 32 || looksLikePlaceholder(sessionSecret) {
		return Config{}, fmt.Errorf("SESSION_SECRET must contain at least 32 non-placeholder characters")
	}

	origins, err := parseOrigins(value("ALLOWED_ORIGINS", ""))
	if err != nil {
		return Config{}, err
	}
	ttl, err := duration("SESSION_TTL", 12*time.Hour)
	if err != nil {
		return Config{}, err
	}
	idle, err := duration("SESSION_IDLE_TIMEOUT", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}
	if idle > ttl {
		return Config{}, fmt.Errorf("SESSION_IDLE_TIMEOUT cannot exceed SESSION_TTL")
	}
	lockout, err := duration("LOGIN_LOCKOUT", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}

	cookieSecure, err := boolean("AUTH_COOKIE_SECURE", appEnv == "production")
	if err != nil {
		return Config{}, err
	}
	if appEnv == "production" && !cookieSecure {
		return Config{}, fmt.Errorf("AUTH_COOKIE_SECURE must be true in production")
	}
	sameSite, err := sameSite(value("AUTH_COOKIE_SAME_SITE", "lax"))
	if err != nil {
		return Config{}, err
	}
	if sameSite == http.SameSiteNoneMode && !cookieSecure {
		return Config{}, fmt.Errorf("AUTH_COOKIE_SAME_SITE=none requires AUTH_COOKIE_SECURE=true")
	}

	cost, err := integer("PASSWORD_BCRYPT_COST", 12, 10, 14)
	if err != nil {
		return Config{}, err
	}
	maxAttempts, err := integer("LOGIN_MAX_ATTEMPTS", 5, 3, 20)
	if err != nil {
		return Config{}, err
	}
	authRate, err := integer("AUTH_RATE_LIMIT_PER_MINUTE", 10, 1, 1000)
	if err != nil {
		return Config{}, err
	}
	apiRate, err := integer("API_RATE_LIMIT_PER_MINUTE", 120, 1, 10000)
	if err != nil {
		return Config{}, err
	}
	autoMigrate, err := boolean("DB_AUTO_MIGRATE", false)
	if err != nil {
		return Config{}, err
	}
	trustProxy, err := boolean("TRUST_PROXY", false)
	if err != nil {
		return Config{}, err
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		value("DB_HOST", "localhost"), value("DB_PORT", "5432"), value("DB_USER", "starter"),
		dbPassword, value("DB_NAME", "starter"), value("DB_SSLMODE", "disable"))

	return Config{
		AppEnv: appEnv, LogLevel: logLevel(value("LOG_LEVEL", defaultLogLevel(appEnv))),
		ServerHost: value("SERVER_HOST", "127.0.0.1"), ServerPort: value("SERVER_PORT", "8080"),
		DatabaseDSN: dsn, AutoMigrate: autoMigrate, AllowedOrigins: origins, TrustProxy: trustProxy,
		SessionSecret: sessionSecret, SessionTTL: ttl, SessionIdleTimeout: idle,
		CookieSecure: cookieSecure, CookieSameSite: sameSite, PasswordBcryptCost: cost,
		LoginMaxAttempts: maxAttempts, LoginLockout: lockout, AuthRatePerMinute: authRate,
		APIRatePerMinute: apiRate,
	}, nil
}

func required(key string) (string, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return v, nil
}

func value(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func boolean(key string, fallback bool) (bool, error) {
	raw := value(key, "")
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", key)
	}
	return v, nil
}

func integer(key string, fallback, min, max int) (int, error) {
	raw := value(key, "")
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < min || v > max {
		return 0, fmt.Errorf("%s must be between %d and %d", key, min, max)
	}
	return v, nil
}

func duration(key string, fallback time.Duration) (time.Duration, error) {
	raw := value(key, "")
	if raw == "" {
		return fallback, nil
	}
	v, err := time.ParseDuration(raw)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return v, nil
}

func parseOrigins(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("ALLOWED_ORIGINS is required")
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimRight(strings.TrimSpace(part), "/")
		if origin == "*" {
			return nil, fmt.Errorf("ALLOWED_ORIGINS cannot contain wildcard origins")
		}
		u, err := url.Parse(origin)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" {
			return nil, fmt.Errorf("ALLOWED_ORIGINS contains an invalid origin")
		}
		out = append(out, origin)
	}
	return out, nil
}

func sameSite(raw string) (http.SameSite, error) {
	switch strings.ToLower(raw) {
	case "strict":
		return http.SameSiteStrictMode, nil
	case "lax":
		return http.SameSiteLaxMode, nil
	case "none":
		return http.SameSiteNoneMode, nil
	default:
		return 0, fmt.Errorf("AUTH_COOKIE_SAME_SITE must be strict, lax, or none")
	}
}

func looksLikePlaceholder(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{"replace", "change-me", "changeme", "example", "placeholder"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func defaultLogLevel(env string) string {
	if env == "production" {
		return "info"
	}
	return "debug"
}

func logLevel(raw string) slog.Level {
	switch strings.ToLower(raw) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
