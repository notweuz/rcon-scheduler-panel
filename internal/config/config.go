package config

import (
	"fmt"
	"path/filepath"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	defaultJWTSecret = "secret"
	minJWTSecretLen  = 32
)

type Config struct {
	DatabaseDSN       string `mapstructure:"DATABASE_DSN" validate:"required"`
	JWTSecret         string `mapstructure:"JWT_SECRET" validate:"required"`
	LogLevel          string `mapstructure:"LOG_LEVEL" validate:"oneof=trace debug info warn error fatal panic disabled"`
	AppPort           int    `mapstructure:"APP_PORT" validate:"min=1,max=65535"`
	Timezone          string `mapstructure:"TIMEZONE" validate:"required"`
	GinMode           string `mapstructure:"GIN_MODE" validate:"oneof=debug release test"`
	AllowRegistration bool   `mapstructure:"ALLOW_REGISTRATION"`
}

func NewConfig(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(filepath.Join(path, ".env"))
	v.SetConfigType("env")
	v.AutomaticEnv()

	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("APP_PORT", 8080)
	v.SetDefault("JWT_SECRET", defaultJWTSecret)
	v.SetDefault("DATABASE_DSN", "host=localhost user=postgres password=postgres dbname=rcon-scheduler-panel port=5432 sslmode=disable")
	v.SetDefault("TIMEZONE", "Europe/Moscow")
	v.SetDefault("GIN_MODE", "debug")
	v.SetDefault("ALLOW_REGISTRATION", false)

	err := v.ReadInConfig()
	if err != nil {
		log.Error().Err(err).Msg("couldn't read in config from file")
	}

	var cfg Config
	err = v.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		return field.Tag.Get("mapstructure")
	})

	if err := validate.Struct(c); err != nil {
		return err
	}

	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return fmt.Errorf("invalid TIMEZONE %q: %w", c.Timezone, err)
	}

	if c.GinMode == gin.ReleaseMode && (c.JWTSecret == defaultJWTSecret || len(c.JWTSecret) < minJWTSecretLen) {
		return fmt.Errorf("JWT_SECRET must be a random string of at least %d characters in release mode", minJWTSecretLen)
	}

	return nil
}
