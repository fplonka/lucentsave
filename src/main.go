package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/pgvector/pgvector-go"

	"gopkg.in/natefinch/lumberjack.v2"

	_ "github.com/lib/pq" // Postgres driver
)

var db *pgxpool.Pool

var postListTemplate *template.Template
var postViewTemplate *template.Template
var signinTemplate *template.Template

// Initialize and parse templates once at startup
func initTemplates() {
	dict := func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	}
	isLast := func(index int, len int) bool {
		return index+1 == len
	}
	getBaseURL := func(rawURL string) (string, error) {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return "", err
		}

		// Extract the host, and remove any port information
		host := parsedURL.Host
		if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
			host = host[:colonIndex]
		}

		return host, nil
	}

	postListTemplate = template.Must(template.New("").
		Funcs(template.FuncMap{"dict": dict, "isLast": isLast, "baseURL": getBaseURL}).
		ParseFiles("templates/posts/postBase.html", "templates/posts/postList.html", "templates/base.html"))

	var err error
	postViewTemplate, err = template.ParseFiles("templates/posts/postBase.html", "templates/posts/postView.html", "templates/base.html")
	if err != nil {
		panic(err)
	}

	signinTemplate, err = template.ParseFiles("templates/signin.html", "templates/base.html")
	if err != nil {
		panic(err)
	}

}

func initDatabase() {
	// Init db connection
	ctx := context.Background()
	var err error
	db, err = pgxpool.New(ctx, os.Getenv("LS2_DB_URL"))
	if err != nil {
		panic(fmt.Errorf("unable to connect to database: %v", err))
	}
	err = db.Ping(ctx)
	if err != nil {
		panic(fmt.Errorf("ping failed: %v", err))
	}
}

func main() {
	initDatabase()
	initTemplates()
	initOpenaiClient()
	// generateEmbeddingsForExistingPosts()

	addHandleFuncs()

	// set up logging
	if os.Getenv("ENV") == "production" {
		logWriter := &lumberjack.Logger{
			Filename:   "log.txt",
			MaxSize:    100, // megabytes after which a new file is created
			MaxBackups: 2,   // number of backups
			MaxAge:     28,  // days
			Compress:   false,
		}
		defer logWriter.Close()

		handler := slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
		slog.SetDefault(slog.New(handler))
	} else {
		// don't print the time in dev env, it makes the logs visually noisier
		ReplaceAttr := func(group []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.Attr{}
			}
			return slog.Attr{Key: a.Key, Value: a.Value}
		}

		// use go run . | jq '.' to pretty print the json
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ReplaceAttr: ReplaceAttr}))
		slog.SetDefault(logger)
	}

	loggedMux := logRequest(http.DefaultServeMux)
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}
