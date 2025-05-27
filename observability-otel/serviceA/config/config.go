package config

import "github.com/caarlos0/env/v10"

type Config struct {
	WeatherServiceURL        string `json:"weather_service_url" env:"WEATHER_SERVICE_URL" envDefault:"http://weather-fetcher:8081"`
	RequestNameOtel          string `json:"request_name_otel" env:"REQUEST_NAME_OTEL" envDefault:"service-a-validation-cep-request"`
	OtelServiceName          string `json:"otel_service_name" env:"OTEL_SERVICE_NAME" envDefault:"service-a-validation-cep"`
	OtelAppName              string `json:"otel_app_name" env:"OTEL_APP_NAME" envDefault:"otel-challenge"`
	OtelExporterOtlpEndpoint string `json:"otel_exporter_otlp_endpoint" env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"otel-collector:4317"`
	HttpPort                 string `json:"http_port" env:"HTTP_PORT" envDefault:"8080"`
}

func Load() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
