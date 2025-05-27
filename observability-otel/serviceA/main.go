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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Input struct {
	Cep string `json:"cep"`
}

type CepInput struct {
	RequestNameOtel   string
	WeatherServiceURL string
}

func initProvider(ctx context.Context, serviceName, collectorURL, appName string) (func(context.Context) error, error) {
	exporter, err := otlptrace.New(ctx,
		otlptracehttp.NewClient(
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithEndpoint(collectorURL),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("Otel tracer: Could not setup exporter: %w", err)
	}

	resources, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("application", appName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("Otel tracer: Could not setup resources: %w", err)
	}

	tracer := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		sdktrace.WithResource(resources),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracer)

	return exporter.Shutdown, nil
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Greacefully shutdown the service
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Initialize OpenTelemetry
	shutdown, err := initProvider(ctx, cfg.OtelServiceName, cfg.OtelExporterOtlpEndpoint, cfg.OtelAppName)
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
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Context done, shutting down server...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}
	log.Println("Server shut down gracefully")
	log.Println("Exiting application")
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

	response, sbError := internal.Fetch(ctx, cep, ci.WeatherServiceURL)
	if sbError != nil {
		http.Error(w, sbError.Message, sbError.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
