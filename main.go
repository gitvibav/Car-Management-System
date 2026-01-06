package main

import (
	"Car-Management-System/driver"
	"Car-Management-System/middleware"
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
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

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
