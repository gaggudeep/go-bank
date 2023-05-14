package util

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"time"
)

type Environment string

const (
	EnvDev  Environment = "dev"
	EnvTest             = "test"
	EnvProd             = "prod"
)

type Config struct {
	Environment           Environment   `mapstructure:"ENVIRONMENT"`
	DBDriver              string        `mapstructure:"DB_DRIVER"`
	DBURL                 string        `mapstructure:"DB_URL"`
	MigrationURL          string        `mapstructure:"MIGRATION_URL"`
	RedisServerAddress    string        `mapstructure:"REDIS_SERVER_ADDRESS"`
	HTTPServerAddress     string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress     string        `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey     string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	EMAIL_SENDER_NAME     string        `mapstructure:"EMAIL_SENDER_NAME"`
	EMAIL_SENDER_ADDRESS  string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EMAIL_SENDER_PASSWORD string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	TokenAccessDuration   time.Duration `mapstructure:"TOKEN_ACCESS_DURATION"`
	RefreshTokenDuration  time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	CustomValidators      []Validator   `mapstructure:"custom-validators"`
}

type ServerConfig struct {
}

type Validator struct {
	Name string         `mapstructure:"name"`
	Func validator.Func `mapstructure:"func"`
}

var CustomValidators = []Validator{
	{
		Name: "amount",
		Func: IsValidAmount,
	},
	{
		Name: "currency",
		Func: IsValidCurrency,
	},
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigFile(path + "/app.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	config.CustomValidators = CustomValidators

	return
}
