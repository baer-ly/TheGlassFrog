package web

import (
	"encoding/json"
	"os"
)

type Platform struct {
	Name            string `json:"name"`
	URLTemplate     string `json:"url_template"`
	SuccessSelector string `json:"success_selector"`
	FailedText      string `json:"failed_text"`
}

type Config struct {
	Platforms []Platform `json:"platforms"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
