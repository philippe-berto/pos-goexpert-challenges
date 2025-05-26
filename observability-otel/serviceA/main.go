package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/philippe-berto/pos-goexpert-challenges/observability-otel/serviceA/config"
	"github.com/philippe-berto/pos-goexpert-challenges/observability-otel/serviceA/internal"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Input struct {
	Cep string `json:"cep"`
}

type CepInput struct {
	RequestNameOtel   string
	WeatherServiceURL string
}

func initProvider(serviceName, collectorURL string) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := grpc.DialContext(ctx, collectorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to collector: %v", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(traceProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return traceProvider.Shutdown, nil

}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Greacefully shutdown the service
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Initialize OpenTelemetry
	shutdown, err := initProvider(cfg.OtelServiceName, cfg.OtelExporterOtlpEndpoint)
	if err != nil {
		log.Fatalf("Error initializing OpenTelemetry: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down OpenTelemetry: %v", err)
		}
	}()

	// Initialize the service

	ci := CepInput{
		RequestNameOtel:   cfg.RequestNameOtel,
		WeatherServiceURL: cfg.WeatherServiceURL,
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second)) // 60 seconds

	router.Handle("/metrics", promhttp.Handler())
	router.Post("/", ci.cepInput)

	server := http.Server{
		Addr:    ":" + cfg.HttpPort,
		Handler: router,
	}

	go func() {
		log.Printf("Starting server on :%v", cfg.HttpPort)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	select {
	case <-sigCh:
		log.Println("Received shutdown signal, shutting down gracefully...")
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down server: %v", err)
		}
		log.Println("Server shut down gracefully")
	case <-ctx.Done():
		log.Println("Context done, shutting down server...")
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down server: %v", err)
		}
		log.Println("Server shut down gracefully")
	}
	log.Println("Exiting application")
	os.Exit(0)
}

func (ci *CepInput) cepInput(w http.ResponseWriter, req *http.Request) {
	carrier := propagation.HeaderCarrier(req.Header)
	ctx := req.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	tracer := otel.GetTracerProvider().Tracer(ci.RequestNameOtel, trace.WithInstrumentationVersion("semver:"))

	ctx, span := tracer.Start(ctx, ci.RequestNameOtel)
	defer span.End()

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

	fetcher := internal.New(ctx, ci.WeatherServiceURL, carrier)
	response, sbError := fetcher.Fetch(cep)
	if sbError != nil {
		http.Error(w, sbError.Message, sbError.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
