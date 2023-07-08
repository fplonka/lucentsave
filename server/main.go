package main

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/lib/pq" // Postgres driver
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/rs/cors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal().Err(err)
	}

	err = prepareStatements()
	if err != nil {
		log.Fatal().Err(err)
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

	// Logging config
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("ENV") == "production" {
		logWriter := &lumberjack.Logger{
			Filename:   "log.txt",
			MaxSize:    100, // megabytes after which a new file is created
			MaxBackups: 2,   // number of backups
			MaxAge:     28,  // days
			Compress:   false,
		}
		defer logWriter.Close()
		log.Logger = zerolog.New(logWriter).With().Timestamp().Logger()

	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	log.Fatal().Err(http.ListenAndServe("localhost:8080", logRequest(handler)))
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//time.Sleep(time.Second * 3)
		next.ServeHTTP(w, r)
	})
}
