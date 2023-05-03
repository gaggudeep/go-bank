package util

import (
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
	DBConfig     DBConfig     `mapstructure:"db"`
	ServerConfig ServerConfig `mapstructure:"server"`
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
	return
}
