package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const configFileName = "/.gatorconfig.json"

func Read() (Config, error) {
	cfgpath, err := getConfigFilepath()
	if err != nil {
		return Config{}, err
	}
	data, err := ioutil.ReadFile(cfgpath)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

func (config *Config) SetUser(username string) error {
	config.CurrentUserName = username
	err := write(*config)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFilepath() (string, error) {
	usersHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	jsonFile := usersHome + configFileName
	return jsonFile, nil
}

func write(cfg Config) error {
	cfgpath, err := getConfigFilepath()
	if err != nil {
		return err
	}
	file, err := os.Create(cfgpath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}
	return nil
}
