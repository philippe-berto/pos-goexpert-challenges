package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/philippe-berto/pos-goexpert-challenges/cloud-run-deploy/router"
	"github.com/philippe-berto/pos-goexpert-challenges/observability-otel/serviceA/internal"
)

type Input struct {
	Cep string `json:"cep"`
}

func main() {
	ctx := context.Background()

	router := router.New(ctx)

	router.AddRoute("POST", "/", CepInput)

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("Starting server on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func CepInput(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	defer req.Body.Close()

	var input Input
	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	cep := input.Cep
	if err := internal.VerifyCep(cep); err != nil {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	fetcher := internal.New(req.Context(), "http://service-b:8081")
	response, sbError := fetcher.Fetch(cep)
	if sbError != nil {
		http.Error(w, sbError.Message, sbError.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
