package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

func rescrapeEmptyPosts() {
	log.Println("rescrape: querying posts with empty body...")

	rows, err := db.Query(context.Background(),
		`SELECT id, url, user_id FROM posts WHERE body = '' ORDER BY id`)
	if err != nil {
		log.Fatalf("rescrape: query failed: %v", err)
	}
	defer rows.Close()

	type postRef struct {
		id     int
		url    string
		userID int
	}

	var posts []postRef
	for rows.Next() {
		var p postRef
		if err := rows.Scan(&p.id, &p.url, &p.userID); err != nil {
			log.Printf("rescrape: scan failed: %v", err)
			continue
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rescrape: row iteration error: %v", err)
	}

	log.Printf("rescrape: found %d posts to scrape", len(posts))

	succeeded, failed := 0, 0
	placeholder := "<p>We were unable to restore this post's content following a security breach " +
		"at our previous hosting provider. The original URL is preserved above — " +
		"you can visit it to read the article.</p>"

	for i, p := range posts {
		log.Printf("rescrape: [%d/%d] post %d: %s", i+1, len(posts), p.id, p.url)

		title, body, err := fetchArticle(p.url)
		if err != nil {
			log.Printf("rescrape: failed to fetch post %d: %v", p.id, err)
			title = domainFromURL(p.url)
			body = placeholder
			failed++
		} else {
			// node sometimes returns empty title for pages it partially parsed
			if title == "" {
				title = domainFromURL(p.url)
			}
			succeeded++
		}

		_, err = db.Exec(context.Background(),
			`UPDATE posts SET title = $1, body = $2 WHERE id = $3`,
			title, body, p.id)
		if err != nil {
			log.Printf("rescrape: failed to update post %d: %v", p.id, err)
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("rescrape: embedding panicked for post %d: %v", p.id, r)
				}
			}()
			saveEmbedding(Post{ID: p.id, URL: p.url, Title: title, Body: body, UserID: p.userID})
		}()

		// small delay to be nice to node server and openai
		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("rescrape: done. %d succeeded, %d used placeholder.", succeeded, failed)
}

func fetchArticle(articleURL string) (title, content string, err error) {
	reqBody, _ := json.Marshal(map[string]string{"url": articleURL})

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post("http://localhost:3000/process", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", "", fmt.Errorf("node request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("node returned status %d", resp.StatusCode)
	}

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", fmt.Errorf("decode failed: %w", err)
	}

	return data["title"], data["content"], nil
}

func domainFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Hostname()
}
