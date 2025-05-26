package config

import "github.com/caarlos0/env/v10"

type Config struct {
	Title                    string `json:"title" env:"TITLE" envDefault:"Weather Fetcher"`
	Content                  string `json:"content" env:"CONTENT" envDefault:"Service B Weather Fetcher"`
	BackgroundColor          string `json:"background_color" env:"BACKGROUND_COLOR" envDefault:"green"`
	RequestNameOtel          string `json:"request_name_otel" env:"REQUEST_NAME_OTEL" envDefault:"service-b-weather-fetcher-request"`
	OtelServiceName          string `json:"otel_service_name" env:"OTEL_SERVICE_NAME" envDefault:"service-b-weather-fetcher"`
	OtelExporterOtlpEndpoint string `json:"otel_exporter_otlp_endpoint" env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"otel-collector:4317"`
	HttpPort                 string `json:"http_port" env:"HTTP_PORT" envDefault:"8081"`
	WAPI_KEY                 string `json:"wapi_key" env:"WAPI_KEY" envDefault:"a3261b2cece24bacbb8134302252305"`
}

func Load() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
