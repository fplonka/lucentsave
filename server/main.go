package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // Postgres driver

	"github.com/rs/cors"
)

func main() {
	_, _ = un, trace

	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	err = prepareStatements()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	addHandleFuncs(mux)

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:4173", "http://localhost:3000", "http://185.196.220.234:3000", "185.196.220.234:3000", "https://185.196.220.234:3000",
			"http://lucentsave.com:3000", "https://lucentsave.com:3000", "lucentsave.com:3000", "lucentsave.com", "www.lucentsave.com", "http://lucentsave.com", "https://lucentsave.com", "https://www.lucentsave.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	log.Fatal(http.ListenAndServe("localhost:8080", logRequest(handler)))
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		fmt.Println()
		next.ServeHTTP(w, r)
	})
}
