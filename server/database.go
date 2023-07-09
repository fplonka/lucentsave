package main

import (
	"database/sql"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
)

// TODO: update schema
type Post struct {
	// Also has a user_id in the databse
	ID        int    `json:"id"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	IsRead    bool   `json:"isRead"`
	IsLiked   bool   `json:"isLiked"`
	TimeAdded int64  `json:"-"`
}

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"` // TODO: rename this, confusing AF
}

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")

	checkUserIsPostOwnerStmt  *sql.Stmt
	getUserPostsStmt          *sql.Stmt
	getPostStmt               *sql.Stmt
	addPostStmt               *sql.Stmt
	deletePostStmt            *sql.Stmt
	updatePostStatusStmt      *sql.Stmt
	getUserIDByEmailStmt      *sql.Stmt
	getUserHashedPasswordStmt *sql.Stmt
	addUserStmt               *sql.Stmt
	getPostsByBodySearchStmt  *sql.Stmt
)

func prepareStatements() error {
	// TODO: refactor error-checking
	var err error
	checkUserIsPostOwnerStmt, err = db.Prepare("SELECT user_id FROM posts WHERE id = $1")
	if err != nil {
		return err
	}
	getUserPostsStmt, err = db.Prepare("SELECT id, title, read, liked, url FROM posts WHERE user_id = $1 ORDER BY added_at DESC")
	if err != nil {
		return err
	}
	getPostStmt, err = db.Prepare("SELECT id, title, body, read, liked, url FROM posts WHERE id = $1 AND user_id = $2")
	if err != nil {
		return err
	}
	addPostStmt, err = db.Prepare("INSERT INTO posts (user_id, title, body, url, added_at) VALUES ($1, $2, $3, $4, $5) RETURNING id")
	if err != nil {
		return err
	}
	deletePostStmt, err = db.Prepare("DELETE FROM posts WHERE ID = $1")
	if err != nil {
		return err
	}
	updatePostStatusStmt, err = db.Prepare("UPDATE posts SET read = $1, liked = $2 WHERE id = $3 AND user_id = $4")
	if err != nil {
		return err
	}
	getUserIDByEmailStmt, err = db.Prepare("SELECT id FROM users WHERE email = $1")
	if err != nil {
		return err
	}
	getUserHashedPasswordStmt, err = db.Prepare("SELECT hashed_password FROM users WHERE email = $1")
	if err != nil {
		return err
	}
	addUserStmt, err = db.Prepare("INSERT INTO users (email, hashed_password) VALUES ($1, $2) RETURNING id")
	if err != nil {
		return err
	}
	getPostsByBodySearchStmt, err = db.Prepare(`
	SELECT id, title, read, liked, url, ts_rank(
		setweight(to_tsvector('english', title), 'A') || 
		setweight(to_tsvector('english', url), 'B') || 
		setweight(to_tsvector('english', body), 'C'), 
		plainto_tsquery('english', $2)
	) as relevancy
	FROM posts
	WHERE user_id = $1 AND (
		setweight(to_tsvector('english', title), 'A') || 
		setweight(to_tsvector('english', url), 'B') || 
		setweight(to_tsvector('english', body), 'C')
	) @@ plainto_tsquery('english', $2)
	ORDER BY relevancy DESC
`)

	if err != nil {
		return err
	}
	return nil
}

var db *sql.DB

func checkUserIsPostOwner(userID, postID int) error {
	log.Info().Int("userID", userID).Int("postID", postID).Str("query", "checkUserIsPostOwner").Msg("")
	var postOwnerID int
	err := checkUserIsPostOwnerStmt.QueryRow(postID).Scan(&postOwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	if postOwnerID != userID {
		return ErrUnauthorized
	}

	return nil
}

func getUserPosts(userID int) ([]Post, error) {
	log.Info().Int("userID", userID).Str("query", "getUserPosts").Msg("")
	rows, err := getUserPostsStmt.Query(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]Post, 0)

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.IsRead, &post.IsLiked, &post.URL); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func getPost(userID, postID int) (Post, error) {
	err := checkUserIsPostOwner(userID, postID)
	if err != nil {
		return Post{}, err
	}

	var post Post
	log.Info().Int("userID", userID).Int("postID", postID).Str("query", "getPost").Msg("")
	err = getPostStmt.QueryRow(postID, userID).Scan(&post.ID, &post.Title, &post.Body, &post.IsRead, &post.IsLiked, &post.URL)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func addPost(post Post, userID int) (int, error) {
	log.Info().Int("userID", userID).Str("query", "addPost").Msg("")
	// By default, a post will have read and liked set to false
	var id int
	err := addPostStmt.QueryRow(userID, post.Title, post.Body, post.URL, time.Now().Unix()).Scan(&id)

	if err != nil {
		return -1, err
	}
	return id, nil
}

func addUser(user User, hashedPassword []byte) (int, error) {
	log.Info().Str("email", user.Email).Str("query", "addUser").Msg("")
	var id int
	err := addUserStmt.QueryRow(user.Email, hashedPassword).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func deletePost(userID, postID int) error {
	log.Info().Int("userID", userID).Int("postID", postID).Str("query", "checkUserIsPostOwner").Msg("")
	err := checkUserIsPostOwner(userID, postID)
	if err != nil {
		return err
	}

	log.Info().Int("userID", userID).Int("postID", postID).Str("query", "deletePost").Msg("")
	_, err = deletePostStmt.Exec(postID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func updatePostStatus(userID, postID int, read, liked bool) error {
	log.Info().Int("userID", userID).Int("postID", postID).Str("query", "checkUserIsPostOwner").Msg("")
	err := checkUserIsPostOwner(userID, postID)
	if err != nil {
		return err
	}

	log.Info().Int("userID", userID).Int("postID", postID).Str("query", "updatePostStatus").Msg("")
	_, err = updatePostStatusStmt.Exec(read, liked, postID, userID)

	if err != nil {
		return err
	}

	return nil
}

func getIDIfUserExists(email string) (int, error) {
	var id int
	log.Info().Str("email", email).Str("query", "getUserIDByEmail").Msg("")
	err := getUserIDByEmailStmt.QueryRow(email).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func getPostsBySearchInBody(searchString string, userID int) ([]Post, error) {
	log.Info().Int("userID", userID).Str("query", "getPostsByBodySearch").Msg("")
	rows, err := getPostsByBodySearchStmt.Query(userID, searchString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]Post, 0)

	for rows.Next() {
		var post Post
		var relevancy float64
		if err := rows.Scan(&post.ID, &post.Title, &post.IsRead, &post.IsLiked, &post.URL, &relevancy); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil

}

// TODO: right type? string vs byte[] and postgres
func getUserHashedPassword(email string) ([]byte, error) {
	var hashedPassword []byte
	log.Info().Str("email", email).Str("query", "getUserHashedPassword").Msg("")
	err := getUserHashedPasswordStmt.QueryRow(email).Scan(&hashedPassword)
	if err != nil {
		return nil, err
	}
	return hashedPassword, nil
}

// TODO: research indexes, transactions
