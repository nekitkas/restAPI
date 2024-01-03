package apiserver

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port string `json:"port"`
}

func NewConfig() *Config {
	return &Config{
		Port: ":8080",
	}
}

func (c *Config) ReadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, c)
	if err != nil {
		return err
	}

	return nil
}
