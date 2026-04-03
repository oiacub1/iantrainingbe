package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DynamoDB DynamoDBConfig
	AWS      AWSConfig
	App      AppConfig
}

type DynamoDBConfig struct {
	TableName string
	Region    string
	Endpoint  string
}

type AWSConfig struct {
	Region string
}

type AppConfig struct {
	Environment string
	LogLevel    string
	Port        string
	RunLocal    bool
}

func Load() (*Config, error) {
	cfg := &Config{
		DynamoDB: DynamoDBConfig{
			TableName: getEnv("DYNAMODB_TABLE_NAME", "training-platform"),
			Region:    getEnv("AWS_REGION", "us-east-1"),
			Endpoint:  getEnv("DYNAMODB_ENDPOINT", ""),
		},
		AWS: AWSConfig{
			Region: getEnv("AWS_REGION", "us-east-1"),
		},
		App: AppConfig{
			Environment: getEnv("ENVIRONMENT", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
			Port:        getEnv("PORT", "8080"),
			RunLocal:    getEnvAsBool("RUN_LOCAL", true),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DynamoDB.TableName == "" {
		return fmt.Errorf("DYNAMODB_TABLE_NAME is required")
	}

	if c.AWS.Region == "" {
		return fmt.Errorf("AWS_REGION is required")
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
