package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/philippe-berto/pos-goexpert-challenges/cloud-run-deploy/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type ServiceBError struct {
	Message    string
	StatusCode int
}

func Fetch(ctx context.Context, cep, serviceBUrl string) (service.Response, *ServiceBError) {
	payload := map[string]string{
		"cep": cep,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return service.Response{}, &ServiceBError{
			Message:    "failed to marshal payload",
			StatusCode: http.StatusInternalServerError,
		}
	}
	req, err := http.NewRequest(http.MethodPost, serviceBUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return service.Response{}, &ServiceBError{
			Message:    "failed to create request",
			StatusCode: http.StatusInternalServerError,
		}
	}
	req.Header.Set("Content-Type", "application/json")

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return service.Response{}, &ServiceBError{
			Message:    "failed to send request",
			StatusCode: http.StatusInternalServerError,
		}
	}

	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		msg := "unknown error"
		if err == nil {
			msg = string(body)
		}
		return service.Response{}, &ServiceBError{
			Message:    msg,
			StatusCode: res.StatusCode,
		}
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return service.Response{}, &ServiceBError{
			Message:    "failed to read response body",
			StatusCode: http.StatusInternalServerError,
		}
	}

	response := service.Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err)
		return service.Response{}, &ServiceBError{
			Message:    "failed to unmarshal response",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return response, nil
}
