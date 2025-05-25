package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/philippe-berto/pos-goexpert-challenges/multithread/models"
)

const (
	brasilApiTimeout = 5 * time.Second // 5 seconds
	key              = "eab938e385494c25bfe193525251005"
)

type (
	BrasilAPIError struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Name    string `json:"name"`
		Errors  []struct {
			Name    string `json:"name"`
			Message string `json:"message"`
			Service string `json:"service"`
		} `json:"errors"`
	}

	Response struct {
		TempC float64 `json:"temp_C"`
		TempF float64 `json:"temp_F"`
		TempK float64 `json:"temp_K"`
	}

	Current struct {
		TempC float64 `json:"temp_c"`
	}

	WeatherResponse struct {
		Current Current `json:"current"`
	}

	Cep struct {
		ctx        context.Context
		needVerify bool
	}
)

func New(ctx context.Context, needVerify bool) (*Cep, error) {
	return &Cep{
		ctx:        ctx,
		needVerify: needVerify,
	}, nil
}

func (c *Cep) GetWeather(cep string) (Response, error) {
	if c.needVerify {
		if err := c.verifyCep(cep); err != nil {
			log.Println(err)
			return Response{}, fmt.Errorf("WRONG_FORMAT")
		}
	}

	location, err := c.GetFromBrasilCep(cep)
	if err != nil {
		log.Println(err)
		return Response{}, fmt.Errorf("NOT_FOUND")
	}
	temp, err := c.GetTemperature(location.City)
	if err != nil {
		log.Println(err)
		return Response{}, fmt.Errorf("NOT_FOUND")
	}

	return Response{
		TempC: temp.Current.TempC,
		TempF: celsiusToFahrenheit(temp.Current.TempC),
		TempK: celsiusToKelvin(temp.Current.TempC),
	}, nil
}

func (c *Cep) GetLocation(cep string) (string, error) {
	if err := c.verifyCep(cep); err != nil {
		return "", fmt.Errorf("invalid CEP: %s", cep)
	}

	location, err := c.GetFromBrasilCep(cep)
	if err != nil {
		log.Println(err)
		return "", nil
	}
	c.GetTemperature(location.City)

	return location.City, nil

}

func (c *Cep) verifyCep(cep string) error {
	if len(cep) != 8 {
		return fmt.Errorf("invalid CEP: %s", cep)
	}

	for _, char := range cep {
		if char < '0' || char > '9' {
			return fmt.Errorf("invalid CEP: %s", cep)
		}
	}

	return nil
}

func (c *Cep) GetFromBrasilCep(cep string) (models.CepBC, error) {
	ctx, cancel := context.WithTimeout(c.ctx, brasilApiTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://brasilapi.com.br/api/cep/v1/"+cep, nil)
	if err != nil {
		return models.CepBC{}, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	switch {
	case ctx.Err() != nil:
		return models.CepBC{}, fmt.Errorf("TIMEOUT_ERROR")
	case err != nil:
		log.Println(err)
		return models.CepBC{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return models.CepBC{}, err
	}

	cepBC := models.CepBC{}
	err = json.Unmarshal(body, &cepBC)
	if err != nil {
		log.Println(err)
		return models.CepBC{}, err
	}

	if cepBC.City == "" {
		return models.CepBC{}, fmt.Errorf("invalid CEP: %s", cep)
	}

	return cepBC, nil
}

func (c *Cep) GetTemperature(city string) (WeatherResponse, error) {
	encodedCity := url.QueryEscape(city)
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", key, encodedCity)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return WeatherResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return WeatherResponse{}, err
	}

	if res.StatusCode != http.StatusOK {
		return WeatherResponse{}, fmt.Errorf("failed to get weather data: %s", res.Status)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return WeatherResponse{}, err
	}

	weatherResponse := WeatherResponse{}
	err = json.Unmarshal(body, &weatherResponse)
	if err != nil {
		log.Println(err)
		return WeatherResponse{}, err
	}

	return weatherResponse, nil
}

func celsiusToFahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32
}
func celsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}
