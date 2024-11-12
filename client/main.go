package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	TimeoutError = "API_CALL_TIMEOUT"
	TimeoutMS    = 300 * time.Millisecond
)

type Response struct {
	Err   *string `json:"err"`
	Value *string `json:"value"`
}

func main() {
	ctx := context.Background()

	res, err := callServer(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	resJson := Response{}
	err = json.Unmarshal(body, &resJson)
	if err != nil {
		log.Println(err)
		return
	}

	if resJson.Err == nil {
		writeFile(*resJson.Value)
	}

	log.Println("res", string(body))
	log.Println("res.Status", res.Status)
}

func callServer(c context.Context) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(c, TimeoutMS)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
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

func writeFile(valor string) {
	data := fmt.Sprintf("Dólar: %s\n", valor)

	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if _, err := file.WriteString(data); err != nil {
		log.Fatal(err)
	}

	log.Println("File written successfully")
}
