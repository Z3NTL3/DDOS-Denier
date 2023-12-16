package filesystem

import (
	"log"

	"github.com/spf13/viper"
)

func ParseEnv() {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
}
