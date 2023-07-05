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
)

func prepareStatements() {
	checkUserIsPostOwnerStmt, _ = db.Prepare("SELECT user_id FROM posts WHERE id = $1")
	getUserPostsStmt, _ = db.Prepare("SELECT id, title, body, read, liked, url, added_at FROM posts WHERE user_id = $1 ORDER BY added_at DESC")
	getPostStmt, _ = db.Prepare("SELECT id, title, body, read, liked, url FROM posts WHERE id = $1 AND user_id = $2")
	addPostStmt, _ = db.Prepare("INSERT INTO posts (user_id, title, body, url, added_at) VALUES ($1, $2, $3, $4, $5)")
	deletePostStmt, _ = db.Prepare("DELETE FROM posts WHERE ID = $1")
	updatePostStatusStmt, _ = db.Prepare("UPDATE posts SET read = $1, liked = $2 WHERE id = $3 AND user_id = $4")
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

// TODO: prepare statements which will be used
// TODO: research indexes, transactions
