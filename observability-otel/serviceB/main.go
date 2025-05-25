package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/philippe-berto/pos-goexpert-challenges/cloud-run-deploy/config"
	"github.com/philippe-berto/pos-goexpert-challenges/cloud-run-deploy/router"
	"github.com/philippe-berto/pos-goexpert-challenges/cloud-run-deploy/service"
)

type Input struct {
	Cep string `json:"cep"`
}

func main() {
	ctx := context.Background()

	router := router.New(ctx)

	router.AddRoute("POST", "/", getWeather)

	server := http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	log.Println("Starting server on :8081")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func getWeather(w http.ResponseWriter, req *http.Request) {
	cepb, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	var input Input
	if err = json.Unmarshal(cepb, &input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	cep := input.Cep

	config, err := config.LoadConfig()
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}

	weatherService, err := service.New(req.Context(), config.WAPI_KEY, false)
	if err != nil {
		http.Error(w, "failed to create a service", http.StatusInternalServerError)
		return
	}

	response, err := weatherService.GetWeather(cep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)

}
