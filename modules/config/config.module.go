package config

import (
	"fmt"
	"nodes-indexer/modules/config/dto"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

var instance *module

type ConfigModule interface {
	GetConfigService() ConfigService
}

type module struct {
	service ConfigService
}

func NewConfigModule( /*v *validator.Validator*/ ) ConfigModule {
	if instance != nil {
		return instance
	}

	app_env, isExists := os.LookupEnv("APP_ENV")
	if !isExists {
		app_env = "development"
	}
	
	env_file := fmt.Sprintf(".env.%s", app_env)
	if err := godotenv.Load(env_file); err != nil {
		panic(fmt.Sprintf("Failed to load environment file %s: %v", env_file, err))
	}

	config := &dto.Config{}
	if err := env.Parse(config); err != nil {
		panic(fmt.Sprintf("Failed to parse environment variables: %v", err))
	}
	
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	// Initialize the configuration service with default values
	configService := NewConfigService(config)

	instance = &module{configService}
	return instance
}

func (m module) GetConfigService() ConfigService {
	return m.service
}