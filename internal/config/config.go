package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	AWS      AWSConfig
	CORS     CORSConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3BucketName    string
	S3Endpoint      string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type LoggingConfig struct {
	Level            string
	Encoding         string
	OutputPaths      []string
	ErrorOutputPaths []string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}

	config.Server.Port = viper.GetString("SERVER_PORT")
	if config.Server.Port == "" {
		config.Server.Port = fmt.Sprintf("%d", viper.GetInt("server.port"))
	}
	config.Server.ReadTimeout = viper.GetDuration("server.read_timeout")
	config.Server.WriteTimeout = viper.GetDuration("server.write_timeout")
	config.Server.IdleTimeout = viper.GetDuration("server.idle_timeout")

	config.Database.Host = viper.GetString("DB_HOST")
	config.Database.Port = viper.GetString("DB_PORT")
	config.Database.User = viper.GetString("DB_USER")
	config.Database.Password = viper.GetString("DB_PASSWORD")
	config.Database.DBName = viper.GetString("DB_NAME")
	config.Database.SSLMode = viper.GetString("DB_SSL_MODE")
	config.Database.MaxOpenConns = viper.GetInt("database.max_open_conns")
	config.Database.MaxIdleConns = viper.GetInt("database.max_idle_conns")
	config.Database.ConnMaxLifetime = viper.GetDuration("database.conn_max_lifetime")

	config.JWT.Secret = viper.GetString("JWT_SECRET")
	accessTokenDuration := viper.GetString("ACCESS_TOKEN_DURATION")
	if accessTokenDuration != "" {
		duration, err := time.ParseDuration(accessTokenDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid ACCESS_TOKEN_DURATION: %w", err)
		}
		config.JWT.AccessTokenDuration = duration
	}
	refreshTokenDuration := viper.GetString("REFRESH_TOKEN_DURATION")
	if refreshTokenDuration != "" {
		duration, err := time.ParseDuration(refreshTokenDuration)
		if err != nil {
			return nil, fmt.Errorf("invalid REFRESH_TOKEN_DURATION: %w", err)
		}
		config.JWT.RefreshTokenDuration = duration
	}

	config.AWS.Region = viper.GetString("AWS_REGION")
	config.AWS.AccessKeyID = viper.GetString("AWS_ACCESS_KEY_ID")
	config.AWS.SecretAccessKey = viper.GetString("AWS_SECRET_ACCESS_KEY")
	config.AWS.S3BucketName = viper.GetString("S3_BUCKET_NAME")
	config.AWS.S3Endpoint = viper.GetString("S3_ENDPOINT")

	allowedOrigins := viper.GetString("ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		config.CORS.AllowedOrigins = viper.GetStringSlice("ALLOWED_ORIGINS")
	}

	config.Logging.Level = viper.GetString("LOG_LEVEL")
	if config.Logging.Level == "" {
		config.Logging.Level = viper.GetString("logging.level")
	}
	config.Logging.Encoding = viper.GetString("logging.encoding")
	config.Logging.OutputPaths = viper.GetStringSlice("logging.output_paths")
	config.Logging.ErrorOutputPaths = viper.GetStringSlice("logging.error_output_paths")

	return config, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Port == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}
	return nil
}
