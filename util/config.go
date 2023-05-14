package util

import (
	"github.com/gaggudeep/bank_go/const/enum"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Environment          enum.Environment `mapstructure:"ENVIRONMENT"`
	DBDriver             string           `mapstructure:"DB_DRIVER"`
	DBURL                string           `mapstructure:"DB_URL"`
	MigrationURL         string           `mapstructure:"MIGRATION_URL"`
	HTTPServerAddress    string           `mapstructure:"HTTP_SERVER_ADDRESS"`
	GRPCServerAddress    string           `mapstructure:"GRPC_SERVER_ADDRESS"`
	TokenSymmetricKey    string           `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	TokenAccessDuration  time.Duration    `mapstructure:"TOKEN_ACCESS_DURATION"`
	RefreshTokenDuration time.Duration    `mapstructure:"REFRESH_TOKEN_DURATION"`
	CustomValidators     []Validator      `mapstructure:"custom-validators"`
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
