package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Cloudinary CloudinaryConfig
	R2         R2Config
}

type ServerConfig struct {
	Port           string   `mapstructure:"PORT"`
	Env            string   `mapstructure:"ENV"`
	AllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
}

type DatabaseConfig struct {
	URL       string `mapstructure:"DATABASE_URL"`
	ListenURL string `mapstructure:"DATABASE_LISTEN_URL"`
	MaxConns  int    `mapstructure:"DATABASE_MAX_CONNS"`
	MinConns  int    `mapstructure:"DATABASE_MIN_CONNS"`
}

type JWTConfig struct {
	AccessSecret  string `mapstructure:"JWT_ACCESS_SECRET"`
	RefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
}

type CloudinaryConfig struct {
	CloudName string `mapstructure:"CLD_CLOUD_NAME"`
	APIKey    string `mapstructure:"CLD_API_KEY"`
	APISecret string `mapstructure:"CLD_API_SECRET"`
}

type R2Config struct {
	AccountID       string `mapstructure:"R2_ACCOUNT_ID"`
	AccessKeyID     string `mapstructure:"R2_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"R2_SECRET_ACCESS_KEY"`
	BucketName      string `mapstructure:"R2_BUCKET_NAME"`
}

func NewConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	// Replace dot with underscore in env keys
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Default values
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("DATABASE_MAX_CONNS", 20)
	viper.SetDefault("DATABASE_MIN_CONNS", 2)

	if err := viper.ReadInConfig(); err != nil {
		// Ignore error if .env file is not found, we might be using environment variables directly
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// If the error is something else, we might want to log it or handle it,
			// but for now we proceed as env vars might be set in environment
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg.Server); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg.Database); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg.JWT); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg.Cloudinary); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg.R2); err != nil {
		return nil, err
	}

	return &cfg, nil
}
