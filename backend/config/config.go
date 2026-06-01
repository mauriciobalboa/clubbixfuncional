package config

import (
	"os"
	"time"
)

type Config struct {
	// Server
	Port           string
	Environment    string
	RequestTimeout time.Duration
	MaxRequestSize int64

	// JWT
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration

	// Database
	DBPath string

	// Security
	AllowedOrigins  []string
	RateLimitPerMin int
	MaxFailedLogins int
	LockoutDuration time.Duration

	// Payments
	AppmaxToken    string
	AppmaxOrderURL string

	// Seed — credenciais iniciais (nunca hardcoded)
	AdminEmail    string
	AdminPassword string
	AdminCPF      string
	DemoEmail     string
	DemoPassword  string
	DemoCPF       string
}

func LoadConfig() *Config {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		RequestTimeout: 30 * time.Second,
		MaxRequestSize: 10 << 20, // 10 MB

		JWTSecret:         getEnv("JWT_SECRET", "your-super-secret-key-change-this-in-production"),
		JWTExpiration:     24 * time.Hour,
		RefreshExpiration: 7 * 24 * time.Hour,

		DBPath: getEnv("DB_PATH", "./clubbix.db"),

		AllowedOrigins:  []string{"http://localhost:3000", "http://localhost:8080", "http://localhost:5500", "http://127.0.0.1:5500"},
		RateLimitPerMin: 60,
		MaxFailedLogins: 5,
		LockoutDuration: 15 * time.Minute,

		AppmaxToken:    getEnv("APPMAX_TOKEN", ""),
		AppmaxOrderURL: getEnv("APPMAX_ORDER_URL", "https://api.appmax.com.br/api/v3/order"),

		AdminEmail:    getEnv("ADMIN_EMAIL", ""),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
		AdminCPF:      getEnv("ADMIN_CPF", ""),
		DemoEmail:     getEnv("DEMO_EMAIL", ""),
		DemoPassword:  getEnv("DEMO_PASSWORD", ""),
		DemoCPF:       getEnv("DEMO_CPF", ""),
	}

	// Em produção, usar variáveis de ambiente
	if cfg.Environment == "production" {
		if cfg.JWTSecret == "your-super-secret-key-change-this-in-production" {
			panic("JWT_SECRET não foi configurado em produção!")
		}
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
