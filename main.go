package main

import (
	"evo-cli/cmd"
	"fmt"

	"github.com/spf13/viper"
)

func main() {
	// Setup config
	viper.SetConfigName("evo-cli")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("config file not found in $HOME/.config/evo-cli.yaml")
		} else {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}

	// Execute main program
	cmd.Execute()
}
