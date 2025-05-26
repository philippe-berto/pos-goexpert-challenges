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

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/philippe-berto/pos-goexpert-challenges/cloud-run-deploy/service"
	"github.com/philippe-berto/pos-goexpert-challenges/observability-otel/serviceB/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	Input struct {
		Cep string `json:"cep"`
	}
	WeaterHandler struct {
		RequestNameOtel string
		WAPIUrl         string
		WAPIKey         string
	}
)

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

	wh := &WeaterHandler{
		RequestNameOtel: cfg.RequestNameOtel,
		WAPIKey:         cfg.WAPI_KEY,
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second)) // 60 seconds

	router.Handle("/metrics", promhttp.Handler())

	router.Post("/", wh.getWeather)

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

func (wh *WeaterHandler) getWeather(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	tracer := otel.GetTracerProvider().Tracer(wh.RequestNameOtel, trace.WithInstrumentationVersion("semver:"))

	ctx, span := tracer.Start(ctx, wh.RequestNameOtel)
	defer span.End()

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

	weatherService, err := service.New(req.Context(), wh.WAPIKey, false)
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
