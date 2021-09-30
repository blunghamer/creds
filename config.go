package creds

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Keypassfile   string
	Listenaddress string
}

// ReadConfig via viper
func ReadConfig() (*Config, error) {

	toolName := "creds"

	viper.SetConfigName(toolName)
	viper.AddConfigPath(".")
	viper.AddConfigPath(fmt.Sprintf("/etc/%v/", toolName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%v", toolName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.config/%v", toolName))

	// enable replacing config keys
	viper.SetEnvPrefix(toolName)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Error().Str("configfile", viper.ConfigFileUsed()).Msg("Unable to load config file, file not found")
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
			log.Error().Err(err).Msg("Unable to load config")
		}
	}

	// well do not print the config, will leak passwords to log otherwise
	log.Info().Strs("keys", viper.GetViper().AllKeys())

	log.Info().Str("Using config file:", viper.ConfigFileUsed())

	var conf = Config{}
	err := viper.Unmarshal(&conf)
	if err != nil {
		log.Error().Msg("Unable to unmarshal config")
		return nil, err
	}

	return &conf, nil
}
