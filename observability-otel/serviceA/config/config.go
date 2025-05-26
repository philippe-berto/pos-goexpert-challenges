package config

import "github.com/caarlos0/env/v10"

type Config struct {
	Title                    string `json:"title" env:"TITLE" envDefault:"Validation CEP"`
	Content                  string `json:"content" env:"CONTENT" envDefault:"Service A Validation CEP"`
	BackgroundColor          string `json:"background_color" env:"BACKGROUND_COLOR" envDefault:"green"`
	WeatherServiceURL        string `json:"weather_service_url" env:"WEATHER_SERVICE_URL" envDefault:"http://weather-fetcher:8081"`
	RequestNameOtel          string `json:"request_name_otel" env:"REQUEST_NAME_OTEL" envDefault:"service-a-validation-cep-request"`
	OtelServiceName          string `json:"otel_service_name" env:"OTEL_SERVICE_NAME" envDefault:"service-a-validation-cep"`
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
