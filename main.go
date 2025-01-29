// main.go

// @title Online Song Library
// @version 1.0
// @description API for managing an online song library

// @host localhost:8081
// @BasePath /

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/noctusha/music/connection"
	"github.com/noctusha/music/handlers"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/noctusha/music/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// main is the entry point of the application.
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	repo, err := connection.NewRepository()
	if err != nil {
		log.Fatalf("error initializing repository: %v", err)
	}
	defer repo.Close()

	migrationDir := os.Getenv("MIGRATION_DIR")
	if migrationDir == "" {
		migrationDir = "file://migrations"
	}

	m, err := migrate.New(
		migrationDir,
		os.Getenv("POSTGRES_CONN"),
	)
	if err != nil {
		log.Fatalf("Error initializing migrations: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Error applying migrations: %v", err)
	}

	handler := handlers.NewHandler(repo)

	router := mux.NewRouter()

	// Swagger UI handler
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	router.Methods(http.MethodGet).Path("/api/songs").HandlerFunc(handler.ListSongs)
	router.Methods(http.MethodGet).Path("/api/songs/{song_id}/text").HandlerFunc(handler.GetText)
	router.Methods(http.MethodDelete).Path("/api/songs/{song_id}/delete").HandlerFunc(handler.DeleteSong)
	router.Methods(http.MethodPatch).Path("/api/songs/{song_id}/edit").HandlerFunc(handler.EditSong)
	router.Methods(http.MethodPost).Path("/api/songs/new").HandlerFunc(handler.NewSong)

	fmt.Printf("server is running on port %v\n", os.Getenv("SERVER_ADDRESS"))

	err = http.ListenAndServe(os.Getenv("SERVER_ADDRESS"), router)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}

}
