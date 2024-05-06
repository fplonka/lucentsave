package main

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pgvector/pgvector-go"
)

type Post struct {
	ID        int
	UserID    int
	URL       string
	Title     string
	Body      string
	IsRead    bool
	IsLiked   bool
	TimeAdded int64

	BodyHTML template.HTML
}

func logError(logger *slog.Logger, msg string, err error, attr ...any) {
	args := append([]any{"error", err}, attr...)
	logger.Error(msg, args...)
}

// if read is true gets only read posts, otherwise only unread posts
func getUserPostsInfo(userID int, getReadPosts bool) []Post {
	ctx := context.Background()

	logger := slog.Default().With("func", "getUserPosts", "userID", userID, "getReadPosts", getReadPosts)
	defer logger.Info("query")

	// Query the database
	rows, err := db.Query(ctx, `
    SELECT id, url, title, is_read, is_liked
    FROM posts 
    WHERE user_id = $1 AND is_read = $2 
    ORDER BY time_added DESC`,
		userID, getReadPosts)
	if err != nil {
		logError(logger, "query to get user posts failed", err)
		return []Post{}
	}
	defer rows.Close()

	postEntries := []Post{}
	// Iterate over the row results
	for rows.Next() {
		var postEntry Post
		err := rows.Scan(&postEntry.ID, &postEntry.URL, &postEntry.Title, &postEntry.IsRead, &postEntry.IsLiked)
		if err != nil {
			logError(logger, "query row scan failed", err)
			continue
		}
		postEntries = append(postEntries, postEntry)
	}

	// Check for any error encountered during iteration
	if err = rows.Err(); err != nil {
		logError(logger, "query row iteration error", err)
	}

	return postEntries
}

func searchUserPosts(userID int, query string) []Post {
	ctx := context.Background()

	logger := slog.Default().With("func", "searchUserPosts", "userID", userID, "query", query)
	defer logger.Info("query")

	queryString := `
    SELECT id, url, title, is_read, is_liked, ts_rank_cd(tsvector_content, plainto_tsquery('english', $2)) AS rank
    FROM posts
    WHERE user_id = $1 AND tsvector_content @@ plainto_tsquery('english', $2)
    ORDER BY rank DESC;
`
	// Execute the database query.
	rows, err := db.Query(ctx, queryString, userID, query)
	if err != nil {
		logError(logger, "query to get user posts failed", err)
		return []Post{}
	}
	defer rows.Close()

	// Initialize the slice to store the fetched posts.
	postEntries := []Post{}

	// Iterate through the query results.
	for rows.Next() {
		var postEntry Post
		var rank float32 // don't actually care about this
		err := rows.Scan(&postEntry.ID, &postEntry.URL, &postEntry.Title, &postEntry.IsRead, &postEntry.IsLiked, &rank)
		if err != nil {
			logError(logger, "query row scan failed", err)
			continue
		}
		// Assume that the like checkbox visibility is based on the read status.
		postEntries = append(postEntries, postEntry)
	}

	// Check for any errors encountered during iteration.
	if err = rows.Err(); err != nil {
		logError(logger, "query row iteration error", err)
	}

	return postEntries
}

func searchUserPostsByEmbedding(userID int, queryEmbedding []float32) []Post {
	ctx := context.Background()

	logger := slog.Default().With("func", "searchUserPostsByEmbedding", "userID", userID)
	defer logger.Info("query")

	queryString := `
    SELECT id, url, title, is_read, is_liked
    FROM posts
    WHERE user_id = $1
    ORDER BY (embedding <#> $2)
    LIMIT 20;
    `

	rows, err := db.Query(ctx, queryString, userID, pgvector.NewVector(queryEmbedding))
	if err != nil {
		logError(logger, "query to search user posts failed", err)
		return []Post{}
	}
	defer rows.Close()

	var postEntries []Post

	for rows.Next() {
		var postEntry Post
		err := rows.Scan(&postEntry.ID, &postEntry.URL, &postEntry.Title, &postEntry.IsRead, &postEntry.IsLiked)
		if err != nil {
			logError(logger, "query row scan failed", err)
			continue
		}

		postEntries = append(postEntries, postEntry)
	}

	if err = rows.Err(); err != nil {
		logError(logger, "query row iteration error", err)
	}

	return postEntries
}

func markPostLiked(postID int, isLiked bool) error {
	logger := slog.Default().With("func", "markPostLiked", "postID", postID, "isLiked", isLiked)
	defer logger.Info("query")

	ctx := context.Background() // Acquire a context; in real applications, pass this from higher up the call chain. TODO:

	sql := `UPDATE posts SET is_liked = $2 WHERE id = $1`
	commandTag, err := db.Exec(ctx, sql, postID, isLiked)
	if err != nil {
		logError(logger, "query to mark post liked failed", err)
		return err
	}

	// Check if the query affected any rows.
	if commandTag.RowsAffected() == 0 {
		logger.Warn("no rows affected")
		return fmt.Errorf("no rows affected, check if the post with ID %d exists", postID)
	}

	return nil
}
func markPostRead(postID int, isRead bool) error {
	logger := slog.Default().With("func", "markPostRead", "postID", postID, "isRead", isRead)
	defer logger.Info("query")

	ctx := context.Background()

	sql := `UPDATE posts SET is_read = $2 WHERE id = $1`
	commandTag, err := db.Exec(ctx, sql, postID, isRead)
	if err != nil {
		logError(logger, "query to mark post liked failed", err)
	}

	// Check if the query affected any rows.
	if commandTag.RowsAffected() == 0 {
		logger.Warn("no rows affected")
		return fmt.Errorf("no rows affected")
	}

	return nil
}

// updatePostStatus updates both the read and liked status of a post given its ID.
func updatePostStatus(postID, userID int, isRead, isLiked bool) error {
	logger := slog.Default().With("func", "updatePostStatus", "postID", postID, "isLiked", isLiked, "isRead", isRead)
	defer logger.Info("query")

	ctx := context.Background() // In a real application, pass context from higher up.

	var sql string
	var err error
	var commandTag pgconn.CommandTag

	// If isRead is false, ensure isLiked is also set to false regardless of the isLiked input.
	if !isRead {
		sql = `UPDATE posts SET is_read = false, is_liked = false WHERE id = $1 AND user_id = $2`
		commandTag, err = db.Exec(ctx, sql, postID, userID)
	} else {
		sql = `UPDATE posts SET is_read = $2, is_liked = $3 WHERE id = $1 AND user_id = $4`
		commandTag, err = db.Exec(ctx, sql, postID, isRead, isLiked, userID)
	}

	if err != nil {
		logError(logger, "query to update post status failed", err)
		return err
	}

	// Check if the query affected any rows.
	if commandTag.RowsAffected() == 0 {
		logger.Warn("no rows affected")
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func getHashedPasswordAndUserId(email string) (string, int, error) {
	logger := slog.Default().With("func", "getHashedPasswordAndUserId", "email", email)
	defer logger.Info("query")

	ctx := context.Background() // TODO?

	// SQL query to fetch the hashed password for a specific email
	sql := `SELECT id, password_hash FROM users WHERE email = $1`

	var userID int
	var hashedPassword string

	err := db.QueryRow(ctx, sql, email).Scan(&userID, &hashedPassword)
	if err != nil {
		logError(logger, "query row failed", err)
		return "", 0, err
	}

	return hashedPassword, userID, nil
}

func getPostContent(postID int, userID int) (Post, error) {
	logger := slog.Default().With("func", "getPostContent", "postID", postID, "userID", userID)
	defer logger.Info("query")

	ctx := context.Background()

	sql := `SELECT id, url, title, body, is_read, is_liked FROM posts WHERE id = $1 AND user_id = $2`
	row := db.QueryRow(ctx, sql, postID, userID)

	var post Post
	var bodyStr string

	err := row.Scan(&post.ID, &post.URL, &post.Title, &bodyStr, &post.IsRead, &post.IsLiked)
	if err != nil {
		logError(logger, "row scan failed", err)
		return Post{}, err
	}
	post.BodyHTML = template.HTML(bodyStr)

	return post, nil
}

func savePost(post Post) (int, error) {
	logger := slog.Default().With("func", "savePost", "url", post.URL)
	defer logger.Info("query")

	ctx := context.Background()

	sql := `INSERT INTO posts (url, title, body, is_read, is_liked, time_added, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var id int // returned id
	err := db.QueryRow(ctx, sql, post.URL, post.Title, post.Body, post.IsRead, post.IsLiked, post.TimeAdded, post.UserID).Scan(&id)
	if err != nil {
		logError(logger, "query row failed", err)
		return 0, err
	}

	return id, nil
}

func deletePost(userID int, postID int) error {
	logger := slog.Default().With("func", "deletePost", "userID", userID, "postID", postID)
	defer logger.Info("query")

	ctx := context.Background()

	sql := `DELETE FROM posts WHERE id = $1 AND user_id = $2`

	// Execute the deletion
	result, err := db.Exec(ctx, sql, postID, userID)
	if err != nil {
		logError(logger, "query execution failed", err)
		return err
	}

	if result.RowsAffected() == 0 {
		logger.Warn("no rows affected")
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func checkUserExists(email string) (bool, error) {
	logger := slog.Default().With("func", "checkUserExists", "email", email)
	defer logger.Info("query")

	ctx := context.Background()

	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := db.QueryRow(ctx, sql, email).Scan(&exists)
	if err != nil {
		logError(logger, "query row failed", err)
		return false, err
	}

	return exists, nil
}

// create user and return id in db
func createUser(email, hashedPassword string) (int, error) {
	logger := slog.Default().With("func", "createUser", "email", email)
	defer logger.Info("query")

	ctx := context.Background()

	sql := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`

	var id int
	err := db.QueryRow(ctx, sql, email, hashedPassword).Scan(&id)
	if err != nil {
		logError(logger, "query row failed", err)
		return 0, err
	}

	return id, nil
}

func setPostEmbedding(postID int, embedding []float32) error {
	logger := slog.Default().With("func", "setPostEmbedding", "postID", postID)
	defer logger.Info("query")

	query := `UPDATE posts SET embedding = $1 WHERE id = $2`
	_, err := db.Exec(context.Background(), query, pgvector.NewVector(embedding), postID)
	if err != nil {
		logError(logger, "query exec failed", err)
		return err
	}
	return nil
}
