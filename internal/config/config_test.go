package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "default_values",
			want: &Config{
				HTTP: struct{ Port int }{Port: 8080},
				Database: struct {
					Host     string
					Port     int
					User     string
					Password string
					DBName   string
					SSLMode  string
				}{
					Host:     "localhost",
					Port:     5434,
					User:     "postgres",
					Password: "postgres",
					DBName:   "pvz_test",
					SSLMode:  "disable",
				},
			},
		},
		{
			name:    "default values",
			envVars: map[string]string{},
			want: &Config{
				HTTP: struct{ Port int }{Port: 8080},
				Database: struct {
					Host     string
					Port     int
					User     string
					Password string
					DBName   string
					SSLMode  string
				}{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					DBName:   "pvz",
					SSLMode:  "disable",
				},
			},
			wantErr: false,
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"HTTP_PORT":   "9090",
				"DB_HOST":     "custom-host",
				"DB_PORT":     "5433",
				"DB_USER":     "custom-user",
				"DB_PASSWORD": "custom-password",
				"DB_NAME":     "custom-db",
				"DB_SSLMODE":  "require",
			},
			want: &Config{
				HTTP: struct{ Port int }{Port: 9090},
				Database: struct {
					Host     string
					Port     int
					User     string
					Password string
					DBName   string
					SSLMode  string
				}{
					Host:     "custom-host",
					Port:     5433,
					User:     "custom-user",
					Password: "custom-password",
					DBName:   "custom-db",
					SSLMode:  "require",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid HTTP_PORT",
			envVars: map[string]string{
				"HTTP_PORT": "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid DB_PORT",
			envVars: map[string]string{
				"DB_PORT": "invalid",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Сохраняем текущие значения переменных окружения
			originalEnv := make(map[string]string)
			for k, v := range tt.envVars {
				if val, exists := os.LookupEnv(k); exists {
					originalEnv[k] = val
				}
				os.Setenv(k, v)
			}

			// Восстанавливаем оригинальные значения после теста
			defer func() {
				for k := range tt.envVars {
					if val, exists := originalEnv[k]; exists {
						os.Setenv(k, val)
					} else {
						os.Unsetenv(k)
					}
				}
			}()

			got, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "existing env variable",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "test-value",
			want:         "test-value",
		},
		{
			name:         "non-existing env variable",
			key:          "NON_EXISTING_KEY",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}
