package configs

import (
	"fmt"
	"github.com/spf13/viper"
)

func Init() error {
	viper.SetConfigName("factory")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Fatal error factory file: %s \n", err)
	}

	viper.SetDefault("general.port", 8080)
	viper.SetDefault("log.level", "info")
	return nil
}
