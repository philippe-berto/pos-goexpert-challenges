package main

import (
	"context"
	"log"
	"net/http"

	mydb "github.com/philippe-berto/pos-goexpert-challenges/client-server-api/server/database"
	httpserver "github.com/philippe-berto/pos-goexpert-challenges/client-server-api/server/http"
	"github.com/philippe-berto/pos-goexpert-challenges/client-server-api/server/quotes/repository"
	"github.com/philippe-berto/pos-goexpert-challenges/client-server-api/server/quotes/usecase"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	File                 string  `env:"DB_FILE" envDefault:"sqlite.s3db"`
	RunMigration         bool    `env:"DB_MIGRATION" envDefault:"true"`
	ApiCallTimeoutMS     int     `env:"API_CALL_TIMEOUT_MS" envDefault:"200"`
	DbOperationTimeoutMS float32 `env:"DB_OPERATION_TIMEOUT_MS" envDefault:"10"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found")
	}

	cfg := Config{}
	err = env.Parse(&cfg)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Printf("config: %+v", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := mydb.New(ctx, mydb.Config{File: cfg.File, RunMigration: cfg.RunMigration})
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.New(ctx, db.GetConnection())
	uHandler := usecase.New(ctx, repo, usecase.Config{ApiCallTimeoutMs: cfg.ApiCallTimeoutMS, DbOperationTimeoutMs: cfg.DbOperationTimeoutMS})
	quoteHandler := httpserver.New(uHandler)
	http.ListenAndServe(":8080", quoteHandler.Router)
}
