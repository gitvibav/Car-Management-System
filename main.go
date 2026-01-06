package main

import (
	"Car-Management-System/driver"
	"Car-Management-System/middleware"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	carHandler "Car-Management-System/handler/car"
	engineHandler "Car-Management-System/handler/engine"
	loginHandler "Car-Management-System/handler/login"
	carService "Car-Management-System/service/car"
	engineService "Car-Management-System/service/engine"
	carStore "Car-Management-System/store/car"
	engineStore "Car-Management-System/store/engine"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	traceProvider, err := startTracing()
	if err != nil {
		log.Fatalf("Failed to start tracing : %v", err)
	}

	defer func() {
		if err := traceProvider.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown the tracing: %v", err)
		}
	}()

	otel.SetTracerProvider(traceProvider)

	driver.InitDB()
	defer driver.CloseDB()

	db := driver.GetDB()
	carStore := carStore.New(db)
	carService := carService.NewCarService(carStore)

	engineStore := engineStore.New(db)
	engineService := engineService.NewEngineService(engineStore)

	carHandler := carHandler.NewCarHandler(carService)
	engineHandler := engineHandler.NewEngineHandler(engineService)

	router := mux.NewRouter()

	router.Use(otelmux.Middleware("Car-Management-System"))
	router.Use(middleware.MetricMiddleware)

	schemaFile := "store/schema.sql"
	if err := executeSchemaFile(db, schemaFile); err != nil {
		log.Fatal("Error while executing the schema file : ", err)
	}

	router.HandleFunc("/login", loginHandler.LoginHandler).Methods("POST")

	protected := router.PathPrefix("/").Subrouter()

	protected.Use(middleware.AuthMiddleware)

	protected.HandleFunc("/cars/{id}", carHandler.GetCarByID).Methods("GET")
	protected.HandleFunc("/cars", carHandler.GetCarByBrand).Methods("GET")
	protected.HandleFunc("/cars", carHandler.CreateCar).Methods("POST")
	protected.HandleFunc("/cars/{id}", carHandler.UpdateCar).Methods("PUT")
	protected.HandleFunc("/cars/{id}", carHandler.DeleteCar).Methods("DELETE")

	protected.HandleFunc("/engine/{id}", engineHandler.GetEngineById).Methods("GET")
	protected.HandleFunc("/engine", engineHandler.CreateEngine).Methods("POST")
	protected.HandleFunc("/engine/{id}", engineHandler.UpdateEngine).Methods("PUT")
	protected.HandleFunc("/engine/{id}", engineHandler.DeleteEngine).Methods("DELETE")

	router.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, router))

}

func executeSchemaFile(db *sql.DB, fileName string) error {
	sqlFile, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(sqlFile))
	if err != nil {
		return err
	}
	return nil
}
func startTracing() (*sdktrace.TracerProvider, error) {
	header := map[string]string{
		"Content-Type": "application/json",
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint("jaeger:4318"),
			otlptracehttp.WithHeaders(header),
			otlptracehttp.WithInsecure(),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("Error Creating new Exporter : %w", err)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay),
		),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("Car-Management-System"),
			),
		),
	)
	return traceProvider, nil
}
