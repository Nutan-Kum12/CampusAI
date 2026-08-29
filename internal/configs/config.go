package configs

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`
	SMTP     SMTPConfig     `mapstructure:"smtp"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
	Env  string `mapstructure:"env"`
}

// IsDevelopment returns true when running in development mode.
// Used by main.go to toggle debug logging and Gin debug mode.
func (s ServerConfig) IsDevelopment() bool {
	return s.Env == "development"
}

// Address returns the "host:port" string for http.Server.
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"ssl_mode"` // "disable" locally, "require" on AWS RDS
}

// DSN returns the PostgreSQL connection string.
// Format: postgres://user:password@host:port/dbname?sslmode=disable
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// JWTConfig holds JSON Web Token settings.
// Two separate secrets so a leaked refresh token can't forge access tokens.
type JWTConfig struct {
	AccessSecret  string        `mapstructure:"access_secret"`
	RefreshSecret string        `mapstructure:"refresh_secret"`
	AccessTTL     time.Duration `mapstructure:"access_ttl"`  // e.g., 15m
	RefreshTTL    time.Duration `mapstructure:"refresh_ttl"` // e.g., 168h (7 days)
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// SMTPConfig holds email provider configuration for sending OTPs.
type SMTPConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	FromEmail string `mapstructure:"from_email"`
}

// Load reads configuration from config.yaml + env vars, then validates.

func Load(configPath string) (*Config, error) {
	// Load .env file if present (local development convenience only).
	// In Docker/AWS, env vars are injected by the orchestrator — no .env needed.
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configPath) // e.g., "config" directory
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// CAMPUSAI_DATABASE_HOST overrides database.host, etc.
	v.SetEnvPrefix("CAMPUSAI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Defaults — the app starts even without a config.yaml file
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.env", "development")

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "campusai")
	v.SetDefault("database.password", "campusai")
	v.SetDefault("database.name", "campusai")
	v.SetDefault("database.ssl_mode", "disable")

	v.SetDefault("jwt.access_secret", "change-me-before-production")
	v.SetDefault("jwt.refresh_secret", "change-me-before-production")
	v.SetDefault("jwt.access_ttl", 15*time.Minute)
	v.SetDefault("jwt.refresh_ttl", 168*time.Hour) // 7 days

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// SMTP Email defaults
	v.SetDefault("smtp.host", "smtp.gmail.com")
	v.SetDefault("smtp.port", 587)
	v.SetDefault("smtp.username", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from_email", "noreply@campusai.edu")

	// Read config file — silently OK if file doesn't exist (defaults + env vars take over)
	if err := v.ReadInConfig(); err != nil {
		if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
			return nil, fmt.Errorf("config file parse error: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Env == "" {
		return fmt.Errorf("server.env must not be empty")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("jwt.access_secret must not be empty")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("jwt.refresh_secret must not be empty")
	}
	if c.JWT.AccessTTL == 0 {
		return fmt.Errorf("jwt.access_ttl must be > 0")
	}
	if c.JWT.RefreshTTL == 0 {
		return fmt.Errorf("jwt.refresh_ttl must be > 0")
	}
	return nil
}
