package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
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
	connString := fmt.Sprintf("postgres://ls2user:%s@localhost/lucentsave2", os.Getenv("LS2_DB_PASSWORD"))
	var err error
	db, err = pgxpool.New(ctx, connString)
	if err != nil {
		panic(fmt.Errorf("Unable to connect to database: %v", err))
	}
}

func copyUsers(db *sql.DB) {
	// Query to select all users
	rows, err := db.Query("SELECT id, email, hashed_password FROM users")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	type user struct {
		id       int
		email    string
		password string
	}

	// Iterate over the results
	users := []user{}
	for rows.Next() {
		var usr user
		if err := rows.Scan(&usr.id, &usr.email, &usr.password); err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}
		// Print the email
		fmt.Println(usr.email)
		users = append(users, usr)
	}
	if err = rows.Err(); err != nil {
		log.Fatalf("Row iteration error: %v", err)
	}

	sort.Slice(users, func(i, j int) bool { return users[i].id < users[j].id })

	for _, usr := range users {
		fmt.Println("USER:", usr.id, usr.email)
		id, err := createUser(usr.email, usr.password)
		if err != nil {
			panic(err)
		}

		fmt.Println("created with", id, "for", usr.id)
	}

}

func migrate() {
	dbOld, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	copyUsers(dbOld)

	if err != nil {
		log.Fatalf("Failed to connect to old database: %v", err)
	}
	defer dbOld.Close()

	// Query to select all posts
	rows, err := dbOld.Query("SELECT id, user_id, title, body, read, liked, url, added_at FROM posts")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	// Iterate over the results
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Body, &post.IsRead, &post.IsLiked, &post.URL, &post.TimeAdded); err != nil {
			log.Printf("Row scan failed: %v", err)
			continue
		}

		post.UserID -= 34
		savePost(post)
		// Print the title
		fmt.Println(post.Title)
	}

	// Check for any errors encountered during iteration
	if err = rows.Err(); err != nil {
		log.Fatalf("Row iteration error: %v", err)
	}

}

func main() {
	initDatabase()
	initTemplates()
	initOpenaiClient()
	// migrate()
	generateEmbeddingsForExistingPosts()

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
