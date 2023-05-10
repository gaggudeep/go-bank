package util

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBUrl                string        `mapstructure:"DB_URL"`
	ServerAddress        string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	TokenAccessDuration  time.Duration `mapstructure:"TOKEN_ACCESS_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	CustomValidators     []Validator   `mapstructure:"custom-validators"`
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
