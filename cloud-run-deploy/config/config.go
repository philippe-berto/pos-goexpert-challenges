package config

import "github.com/caarlos0/env/v10"

type Config struct {
	WAPI_KEY string `json:"wapi_key" envDefault:"a3261b2cece24bacbb8134302252305"`
}

func LoadConfig() (*Config, error) {
	envConfig := &Config{}
	if err := env.Parse(envConfig); err != nil {
		return nil, err
	}
	return envConfig, nil
}
