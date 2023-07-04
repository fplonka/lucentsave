package main

import (
	"database/sql"
	"errors"
	"sort"
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
)

var db *sql.DB

func checkUserIsPostOwner(userID, postID int) error {
	var postOwnerID int
	err := db.QueryRow("SELECT user_id FROM posts WHERE id = $1", postID).Scan(&postOwnerID)
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
	rows, err := db.Query("SELECT id, title, body, read, liked, url, added_at FROM posts WHERE user_id = $1", userID)
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

	sort.Slice(posts, func(i, j int) bool { return posts[i].TimeAdded > posts[j].TimeAdded })

	return posts, nil
}

func getPost(userID, postID int) (Post, error) {
	err := checkUserIsPostOwner(userID, postID)
	if err != nil {
		return Post{}, err
	}

	var post Post
	err = db.QueryRow("SELECT id, title, body, read, liked, url FROM posts WHERE id = $1", postID).
		Scan(&post.ID, &post.Title, &post.Body, &post.IsRead, &post.IsLiked, &post.URL)
	if err != nil {
		return Post{}, err
	}

	return post, nil
}

func addPost(post Post, userID int) error {
	// By default, a post will have read and liked set to false
	_, err := db.Exec("INSERT INTO posts (user_id, title, body, url, added_at) VALUES ($1, $2, $3, $4, $5)", userID, post.Title, post.Body, post.URL, time.Now().Unix())
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

	// Also using the userID so that a user can't delete someone else's post
	_, err = db.Exec("DELETE FROM posts WHERE ID = $1", postID)
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

	sqlStatement := `
	UPDATE posts
	SET read = $2, liked = $3
	WHERE id = $1;`
	_, err = db.Exec(sqlStatement, postID, read, liked)
	if err != nil {
		return err
	}

	return nil
}

// TODO: prepare statements which will be used
// TODO: research indexes, transactions
