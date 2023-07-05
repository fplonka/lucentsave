package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq" // Postgres driver

	"github.com/rs/cors"
)

func main() {
	_, _ = un, trace

	var err error
	db, err = sql.Open("postgres", "user=postgres dbname=mydatabase sslmode=disable")
	prepareStatements()

	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	addHandleFuncs(mux)

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:4173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	log.Fatal(http.ListenAndServe(":8080", handler))
}
