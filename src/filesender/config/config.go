package config

import (
	"fmt"

	helper "filesender/utils"

	viper "github.com/spf13/viper"
)

// ##### Structs ##############################################################

// Config holds configuration data for the application
type Config struct {
	Salt           string
	PasswordHash   string
	EncryptedKey   string
	EncryptedKeyIv string
}

// ##### Methods ##############################################################

// Load loads the configuration data from the config file
func (c *Config) Initialise() {

	viper.SetConfigType("toml")
	viper.SetConfigName("crypto")
	viper.AddConfigPath("./")

}

// Load loads the configuration data from the config file
func (c *Config) Load() {

	err := viper.ReadInConfig()
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Error reading crypto config file: %v", err))
	}

	c.Salt = viper.GetString("salt")
	c.PasswordHash = viper.GetString("password_hash")
	c.EncryptedKey = viper.GetString("encrypted_key")
	c.EncryptedKeyIv = viper.GetString("encrypted_key_iv")
}

// Save loads the configuration data from the config file
func (c *Config) Save() {

	viper.Set("salt", c.Salt)
	viper.Set("password_hash", c.PasswordHash)
	viper.Set("encrypted_key", c.EncryptedKey)
	viper.Set("encrypted_key_iv", c.EncryptedKeyIv)

	err := viper.WriteConfigAs("crypto.toml")
	if err != nil {
		helper.OutputAndExit(fmt.Sprintf("Error writing crypto config file: %v", err))
	}
}
