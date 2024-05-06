package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

var client *openai.Client

func initOpenaiClient() {
	apiKey := os.Getenv("LS2_OPENAI_KEY")
	client = openai.NewClient(apiKey)
}

const maxCharsPerChunk = 16384

func normalize(vec []float32) {
	sum := 0.0
	for _, v := range vec {
		sum += float64(v * v)
	}
	sum = math.Sqrt(sum)
	for i := range vec {
		vec[i] = vec[i] / float32(sum)
	}
}

func getEmbedding(content string) ([]float32, error) {
	chunks := splitIntoChunks(content, maxCharsPerChunk)
	var combinedEmbedding []float32

	for _, chunk := range chunks {
		resp, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
			Model: openai.SmallEmbedding3,
			Input: chunk,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding: %w", err)
		}

		if len(resp.Data) == 0 {
			return nil, fmt.Errorf("no embedding returned")
		}

		embedding := resp.Data[0].Embedding
		if combinedEmbedding == nil {
			combinedEmbedding = make([]float32, len(embedding))
		}

		for i := range combinedEmbedding {
			combinedEmbedding[i] += embedding[i]
		}
	}

	normalize(combinedEmbedding)

	return combinedEmbedding, nil
}

func splitIntoChunks(s string, chunkSize int) []string {
	var chunks []string
	for start := 0; start < len(s); start += chunkSize {
		end := start + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[start:end])
	}
	return chunks
}

func saveEmbedding(post Post) {
	errs := make(chan error, 3)

	var titleEmbedding []float32
	var urlEmbedding []float32
	var bodyEmbedding []float32

	go func() {
		var err error
		titleEmbedding, err = getEmbedding(post.Title)
		errs <- err
	}()
	go func() {
		var err error
		urlEmbedding, err = getEmbedding(post.URL)
		errs <- err
	}()
	go func() {
		var err error
		bodyEmbedding, err = getEmbedding(post.Body)
		errs <- err
	}()

	for range 3 {
		err := <-errs
		if err != nil {
			slog.Error("failed to get post embedding", "error", err)
			return
		}
	}

	embedding := make([]float32, len(bodyEmbedding))
	for i := range embedding {
		embedding[i] = 0.25*titleEmbedding[i] + 0.15*urlEmbedding[i] + 0.6*bodyEmbedding[i]
	}
	normalize(embedding)

	err := setPostEmbedding(post.ID, embedding)
	if err != nil {
		slog.Error("failed to set post embedding", "error", err)
		return
	} else {
		slog.Info("saved post embedding", "postID", post.ID)
	}
}

func generateEmbeddingsForExistingPosts() error {
	query := `
    SELECT id, url, title, body
    FROM posts;
    `

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.URL, &post.Title, &post.Body)
		if err != nil {
			log.Printf("query row scan failed: %v\n", err)
			continue
		}

		saveEmbedding(post)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %w", err)
	}

	return nil
}
