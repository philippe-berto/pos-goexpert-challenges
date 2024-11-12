package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	q "server/quotes"
	"server/quotes/repository"
	"time"
)

const (
	TimeoutError = "EXTERNAL_API_CALL_TIMEOUT"
)

type (
	GetterDollarQuote interface {
		GetDollarQuote() (*string, error)
	}

	Config struct {
		ApiCallTimeoutMs     int     `env:"API_CALL_TIMEOUT_MS" envDefault:"200"`
		DbOperationTimeoutMs float32 `env:"DB_OPERATION_TIMEOUT_MS" envDefault:"10"`
	}

	Usecase struct {
		ctx  context.Context
		repo repository.CreaterDollarQuote
		cfg  Config
	}
)

func New(ctx context.Context, repo repository.CreaterDollarQuote, cfg Config) *Usecase {
	return &Usecase{
		ctx:  ctx,
		repo: repo,
		cfg:  cfg,
	}
}

func (u *Usecase) GetDollarQuote() (*string, error) {
	res, err := apiCall(u.ctx, time.Duration(u.cfg.ApiCallTimeoutMs)*time.Millisecond)
	if err != nil {
		return nil, err
	}

	log.Println("res", res.Status)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var dollarQuotes []q.DollarQuote
	err = json.Unmarshal(body, &dollarQuotes)
	if err != nil {
		return nil, err
	}

	if len(dollarQuotes) == 0 {
		return nil, fmt.Errorf("no quotes found")
	}

	err = u.repo.CreateDollarQuote(u.ctx, dollarQuotes[0], time.Duration(u.cfg.DbOperationTimeoutMs)*time.Millisecond)
	if err != nil {
		return nil, err
	}

	result := dollarQuotes[0].Bid

	return &result, nil
}

func apiCall(c context.Context, t time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(c, t)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	switch {
	case ctx.Err() != nil:
		return nil, errors.New(TimeoutError)
	case err != nil:
		return nil, err
	}

	return resp, nil
}
