package main

import (
	"database/sql"
	"errors"
	"time"
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
	ID             int    `json:"id"`
	Username       string `json:"username"`
	HashedPassword string `json:"password"` // TODO: rename this, confusing AF
}

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")

	checkUserIsPostOwnerStmt *sql.Stmt
	getUserPostsStmt         *sql.Stmt
	getPostStmt              *sql.Stmt
	addPostStmt              *sql.Stmt
	deletePostStmt           *sql.Stmt
	updatePostStatusStmt     *sql.Stmt
	getUserByUsername        *sql.Stmt
	getUserHashedPassword    *sql.Stmt
	addUser                  *sql.Stmt
)

func prepareStatements() error {
	// TODO: refactor error-checking
	var err error
	checkUserIsPostOwnerStmt, err = db.Prepare("SELECT user_id FROM posts WHERE id = $1")
	if err != nil {
		return err
	}
	getUserPostsStmt, err = db.Prepare("SELECT id, title, body, read, liked, url, added_at FROM posts WHERE user_id = $1 ORDER BY added_at DESC")
	if err != nil {
		return err
	}
	getPostStmt, err = db.Prepare("SELECT id, title, body, read, liked, url FROM posts WHERE id = $1 AND user_id = $2")
	if err != nil {
		return err
	}
	addPostStmt, err = db.Prepare("INSERT INTO posts (user_id, title, body, url, added_at) VALUES ($1, $2, $3, $4, $5)")
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
	getUserByUsername, err = db.Prepare("SELECT email FROM users WHERE email = $1")
	if err != nil {
		return err
	}
	getUserHashedPassword, err = db.Prepare("SELECT id, hashed_password FROM users WHERE email = $1")
	if err != nil {
		return err
	}
	addUser, err = db.Prepare("INSERT INTO users (email, hashed_password) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	return nil
}

var db *sql.DB

func checkUserIsPostOwner(userID, postID int) error {
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
	rows, err := getUserPostsStmt.Query(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]Post, 0)

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Body, &post.IsRead, &post.IsLiked, &post.URL, &post.TimeAdded); err != nil {
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
	// err = db.QueryRow("SELECT id, title, body, read, liked, url FROM posts WHERE id = $1", postID).
	err = getPostStmt.QueryRow(postID, userID).Scan(&post.ID, &post.Title, &post.Body, &post.IsRead, &post.IsLiked, &post.URL)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func addPost(post Post, userID int) error {
	// By default, a post will have read and liked set to false
	_, err := addPostStmt.Exec(userID, post.Title, post.Body, post.URL, time.Now().Unix())

	if err != nil {
		return err
	}
	return nil
}

func deletePost(userID, postID int) error {
	err := checkUserIsPostOwner(userID, postID)
	if err != nil {
		return err
	}

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
	err := checkUserIsPostOwner(userID, postID)
	if err != nil {
		return err
	}

	_, err = updatePostStatusStmt.Exec(read, liked, postID, userID)

	if err != nil {
		return err
	}

	return nil
}

func checkUsernameExists(username string) bool {
	var existingUser User
	err := getUserByUsername.QueryRow(username).Scan(&existingUser.Username)
	// Slightly misleading: returns true if username is not taken but some other error occurs. Should be rare
	return err != sql.ErrNoRows
}

// TODO: research indexes, transactions
