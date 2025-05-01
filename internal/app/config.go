package app

// Config представляет конфигурацию приложения
type Config struct {
	Server struct {
		HTTP struct {
			Host string
			Port int
		}
		GRPC struct {
			Host string
			Port int
		}
	}
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string
	}
	JWT struct {
		Secret     string
		Expiration string
	}
	Logging struct {
		Level  string
		Format string
	}
}
