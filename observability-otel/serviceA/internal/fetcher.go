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

type (
	Fetcher struct {
		ctx         context.Context
		serviceBUrl string
		carrier     propagation.HeaderCarrier
	}
	ServiceBError struct {
		Message    string
		StatusCode int
	}
)

func New(ctx context.Context, serviceBUrl string, carrier propagation.HeaderCarrier) *Fetcher {
	return &Fetcher{
		ctx:         ctx,
		serviceBUrl: serviceBUrl,
		carrier:     carrier,
	}
}

func (f *Fetcher) Fetch(cep string) (service.Response, *ServiceBError) {
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
	req, err := http.NewRequest(http.MethodPost, f.serviceBUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return service.Response{}, &ServiceBError{
			Message:    "failed to create request",
			StatusCode: http.StatusInternalServerError,
		}
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	otel.GetTextMapPropagator().Inject(f.ctx, f.carrier)

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
