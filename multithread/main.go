package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	cepValue         = "22461000"
	brasilApiTimeout = 1 * time.Second
	viaCepTimeout    = 1 * time.Second
)

type (
	CepVC struct {
		Cep         string `json:"cep"`
		Logradouro  string `json:"logradouro"`
		Complemento string `json:"complemento"`
		Unidade     string `json:"unidade"`
		Bairro      string `json:"bairro"`
		Localidade  string `json:"localidade"`
		Uf          string `json:"uf"`
		Estado      string `json:"estado"`
		Regiao      string `json:"regiao"`
		Ibge        string `json:"ibge"`
		Gia         string `json:"gia"`
		Ddd         string `json:"ddd"`
		Siafi       string `json:"siafi"`
	}

	CepBC struct {
		Cep          string `json:"cep"`
		State        string `json:"state"`
		City         string `json:"city"`
		Neighborhood string `json:"neighborhood"`
		Street       string `json:"street"`
		Service      string `json:"service"`
	}

	result struct {
		cepBC  CepBC
		cepVC  CepVC
		source string
		err    error
	}
)

func main() {
	c := context.Background()

	ch := make(chan result)

	go getFromViaCep(c, cepValue, ch)
	go getFromBrasilCep(c, cepValue, ch)

	select {
	case result := <-ch:
		close(ch)
		if result.err != nil {
			log.Println(result.err)
		}

		if result.source == "Via Cep" {
			log.Println(result.source)
			jsonData, err := json.Marshal(result.cepVC)
			if err != nil {
				log.Println("Error encoding JSON:", err)
			} else {
				log.Println(string(jsonData))
			}
		}

		if result.source == "Brasil API" {
			log.Println(result.source)
			jsonData, err := json.Marshal(result.cepBC)
			if err != nil {
				log.Println("Error encoding JSON:", err)
			} else {
				log.Println(string(jsonData))
			}
		}

	case <-time.After(1 * time.Second):
		log.Println("Timeout: no response received within 1 second")
	}
}

func getFromViaCep(c context.Context, cep string, ch chan<- result) {
	ctx, cancel := context.WithTimeout(c, viaCepTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://viacep.com.br/ws/"+cep+"/json", nil)
	if err != nil {
		log.Println(err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	switch {
	case ctx.Err() != nil:
		log.Println("TIMEOUT_ERROR")
		return
	case err != nil:
		log.Println(err)
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	cepVC := CepVC{}
	err = json.Unmarshal(body, &cepVC)
	if err != nil {
		log.Println(err)
		return
	}

	ch <- result{cepVC: cepVC, source: "Via Cep", err: nil}
}

func getFromBrasilCep(c context.Context, cep string, ch chan<- result) {
	ctx, cancel := context.WithTimeout(c, brasilApiTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://brasilapi.com.br/api/cep/v1/"+cep, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &http.Client{}
	res, err := client.Do(req)
	switch {
	case ctx.Err() != nil:
		log.Println("TIMEOUT_ERROR")
		return
	case err != nil:
		log.Println(err)
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	cepBC := CepBC{}
	err = json.Unmarshal(body, &cepBC)
	if err != nil {
		log.Println(err)
		return
	}

	ch <- result{cepBC: cepBC, source: "Brasil API", err: nil}
}
