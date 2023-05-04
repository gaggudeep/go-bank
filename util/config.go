package util

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type DBConfig struct {
	Driver string `mapstructure:"driver"`
	URL    string `mapstructure:"url"`
}

type ServerConfig struct {
	Addr string `mapstructure:"address"`
}

type Config struct {
	DBConfig         DBConfig     `mapstructure:"db"`
	ServerConfig     ServerConfig `mapstructure:"server"`
	CustomValidators []Validator  `mapstructure:"custom-validators"`
}

type Validator struct {
	Name string         `mapstructure:"name"`
	Func validator.Func `mapstructure:"func"`
}

var customValidators = []Validator{
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
	viper.SetConfigFile(path + "/app.yml")
	viper.SetConfigType("yml")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	config.CustomValidators = customValidators

	return
}
