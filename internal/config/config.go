package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerConfig ServerConfig
	DBConfig     DBConfig
	JWTConfig    JWTConfig
	Prometheus   PrometheusConfig
	CacheConfig  CacheConfig
}

type ServerConfig struct {
	Port           int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxConnections int
}

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type PrometheusConfig struct {
	Port int
	Path string
}

type CacheConfig struct {
	TTL     time.Duration
	MaxSize int
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, err
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, err
	}

	maxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "1000"))
	if err != nil {
		return nil, err
	}

	maxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "100"))
	if err != nil {
		return nil, err
	}

	readTimeout, err := time.ParseDuration(getEnv("SERVER_READ_TIMEOUT", "50ms"))
	if err != nil {
		return nil, err
	}

	writeTimeout, err := time.ParseDuration(getEnv("SERVER_WRITE_TIMEOUT", "50ms"))
	if err != nil {
		return nil, err
	}

	idleTimeout, err := time.ParseDuration(getEnv("SERVER_IDLE_TIMEOUT", "120s"))
	if err != nil {
		return nil, err
	}

	maxConnections, err := strconv.Atoi(getEnv("SERVER_MAX_CONNECTIONS", "1000"))
	if err != nil {
		return nil, err
	}

	return &Config{
		ServerConfig: ServerConfig{
			Port:           port,
			ReadTimeout:    readTimeout,
			WriteTimeout:   writeTimeout,
			IdleTimeout:    idleTimeout,
			MaxConnections: maxConnections,
		},
		DBConfig: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            dbPort,
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "pvz"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: time.Minute,
		},
		JWTConfig: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key"),
			Expiration: 24 * time.Hour,
		},
		Prometheus: PrometheusConfig{
			Port: 9000,
			Path: "/metrics",
		},
		CacheConfig: CacheConfig{
			TTL:     5 * time.Minute,
			MaxSize: 10000,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
