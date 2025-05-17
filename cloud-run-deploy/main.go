package main

import (
	"context"
	"log"
	"net/http"

	"github.com/philippe-berto/pos-goexpert-challenges/cloud-run-deploy/handler"
	"github.com/philippe-berto/pos-goexpert-challenges/cloud-run-deploy/router"
)

func main() {
	ctx := context.Background()

	h, err := handler.New(ctx)
	if err != nil {
		log.Fatalf("Error creating handler: %v", err)
	}

	r := router.New(ctx)
	r.AddRoute("GET", "/{cep}", h.GetWeather)

	server := http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	log.Println("Starting server on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
