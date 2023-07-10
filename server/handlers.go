package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sym01/htmlsanitizer"
	"golang.org/x/crypto/bcrypt"

	readability "github.com/go-shiori/go-readability"
)

// TODO: method types everywhere...
func addHandleFuncs(mux *http.ServeMux) {
	// TODO: middlewhere for checking post owner id == user id ??
	mux.HandleFunc("/api/getAllUserPosts", authMiddleware(getSavedPostsHandler))
	mux.HandleFunc("/api/getPost", authMiddleware(getPostHandler))
	mux.HandleFunc("/api/createPostFromURL", authMiddleware(createPostFromURLHandler))
	mux.HandleFunc("/api/deletePost", authMiddleware(deletePostHandler))
	mux.HandleFunc("/api/updatePostStatus", authMiddleware(updatePostStatusHandler))
	mux.HandleFunc("/api/createUser", createUserHandler)
	mux.HandleFunc("/api/signout", authMiddleware(signoutHandler))
	mux.HandleFunc("/api/signin", signinHandler)
	mux.HandleFunc("/api/fetchPage", authMiddleware(fetchPageHandler))
	mux.HandleFunc("/api/searchPosts", authMiddleware(searchPostsHandler))
}

func writeErrorResponse(err error, w http.ResponseWriter) {
	switch err {
	case ErrUnauthorized:
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case ErrNotFound:
		http.Error(w, "Resource not found", http.StatusNotFound)
	default:
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getUserIdFromRequest(r *http.Request) int {
	return r.Context().Value(userIDKey).(int)
}

func writeUserPosts(userID int, w http.ResponseWriter, log zerolog.Logger) error {
	posts, err := getUserPosts(userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get posts from database")
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(posts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode posts")
		return err
	}
	return nil
}

const (
	DEFAULT_SAME_SITE_MODE = http.SameSiteLaxMode
)

// Gets all posts for a given user, regardless of whether they're read or liked
// NOTE: this doesn't return the post bodies! just the "metadata"
func getSavedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)

	log := log.With().Str("endpoint", "/getAllUserPosts").Int("userID", userID).Logger()

	err := writeUserPosts(userID, w, log)
	if err != nil {
		http.Error(w, "Failed to get user posts", http.StatusInternalServerError)
	} else {
		log.Info().Msg("Success")
	}
}

func getPostHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)
	queryParams := r.URL.Query()
	postIDStr := queryParams.Get("id")

	log := log.With().Str("endpoint", "/getPost").Int("userID", userID).Logger()

	if postIDStr == "" {
		log.Warn().Msg("No ID provided")
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		log.Warn().Err(err).Msg("Post ID was not integer")
		http.Error(w, "Post ID must be an integer", http.StatusBadRequest)
		return
	}

	log = log.With().Int("postID", postID).Logger()

	post, err := getPost(userID, postID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get post from database")
		writeErrorResponse(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)

	log.Info().Msg("Success")
}

func createPostFromURLHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)
	url := r.URL.Query().Get("url")
	log := log.With().Str("endpoint", "/createPostFromURL").Int("userID", userID).Str("url", url).Logger()

	article, err := readability.FromURL(url, 10*time.Second)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save article from URL")
		http.Error(w, "Failed to save page", http.StatusBadRequest)
		return
	}

	if article.Title == "" || article.Content == "" {
		log.Warn().Msg("Article title or content was empty after parsing")
		http.Error(w, "Failed to save page", http.StatusBadRequest)
		return
	}

	article.Content, err = htmlsanitizer.SanitizeString(article.Content)
	if err != nil {
		log.Error().Err(err).Msg("Failed to sanitize post body")
		http.Error(w, "Failed to save page", http.StatusInternalServerError)
		return
	}

	maxLen := 200000
	if len(url)+len(article.Title)+len(article.Content) > maxLen {
		log.Warn().Msg("Post too long to save")
		http.Error(w, "Article too long to save", http.StatusInternalServerError)
		return
	}

	postID, err := addPost(Post{Title: article.Title, Body: article.Content, URL: url}, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save post to database")
		http.Error(w, "Failed to save page", http.StatusInternalServerError)
		return
	}

	log = log.With().Int("postID", postID).Logger()

	// Return updated posts list
	err = writeUserPosts(userID, w, log)
	if err != nil {
		http.Error(w, "Failed to get updated posts list", http.StatusInternalServerError)
	} else {
		log.Info().Msg("Success")
	}
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)
	queryParams := r.URL.Query()
	postIDStr := queryParams.Get("id")

	log := log.With().Str("endpoint", "/deletePost").Int("userID", userID).Logger()

	if postIDStr == "" {
		log.Warn().Msg("No ID provided")
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		log.Warn().Err(err).Msg("Post ID was not integer")
		http.Error(w, "Post ID must be an integer", http.StatusBadRequest)
		return
	}

	log = log.With().Int("postID", postID).Logger()

	err = deletePost(userID, postID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete post from database")
		writeErrorResponse(err, w)
		return
	}

	// Write updated posts list to response
	err = writeUserPosts(userID, w, log)
	if err != nil {
		http.Error(w, "Failed to get user posts", http.StatusInternalServerError)
	} else {
		log.Info().Msg("Success")
	}
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User

	log := log.With().Str("endpoint", "/createUser").Logger()

	// Decode the incoming User json
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode user from request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log = log.With().Str("email", user.Email).Logger()

	// Check if email is valid
	_, err = mail.ParseAddress(user.Email)
	if err != nil {
		log.Error().Msg("Failed to parse email")
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	// Check if the user already exists
	_, err = getIDIfUserExists(user.Email)
	if err != sql.ErrNoRows {
		if err == nil {
			// User with that email exists
			log.Warn().Msg("Email already taken")
			http.Error(w, "Email already taken", http.StatusBadRequest)
			return
		} else {
			// Some other error was encountered
			log.Error().Err(err).Msg("Failed to check if email is taken")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	// Insert the user into the database.
	userID, err := addUser(user, hashedPassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add user to databse")
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	log = log.With().Int("userID", userID).Logger()

	err = generateAndSetAuthToken(w, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to authenticate after creating user")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	log.Info().Msg("Success")
}

func signinHandler(w http.ResponseWriter, r *http.Request) {

	var user User

	log := log.With().Str("endpoint", "/signin").Logger()

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode user from request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log = log.With().Str("email", user.Email).Logger()

	userID, err := getIDIfUserExists(user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Err(err).Msg("User with provided email not found")
			http.Error(w, "Account not found", http.StatusBadRequest)
			return
		} else {
			log.Error().Err(err).Msg("Failed to get user by email")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	log = log.With().Int("userID", userID).Logger()

	hashedPassword, err := getUserHashedPassword(user.Email)
	if err != nil {
		log.Error().Err(err).Msg("Account not found after being found in previous check")
		http.Error(w, "Account not found", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(user.Password)); err != nil {
		log.Warn().Err(err).Msg("Incorrect password")
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	err = generateAndSetAuthToken(w, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set token")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	err = writeUserPosts(userID, w, log)
	if err != nil {
		http.Error(w, "Failed to get user posts", http.StatusInternalServerError)
	} else {
		log.Info().Msg("Success")
	}
}

func signoutHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)
	log := log.With().Str("endpoint", "/signout").Int("userID", userID).Logger()
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		MaxAge:   -1, // Delete cookie immediately
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: DEFAULT_SAME_SITE_MODE,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "loggedIn",
		Value:    "",
		MaxAge:   -1,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: DEFAULT_SAME_SITE_MODE,
		Path:     "/",
	})
	w.WriteHeader(http.StatusOK)
	log.Info().Msg("Success")
}

type PostUpdate struct {
	Id    int  `json:"id"`
	Read  bool `json:"read"`
	Liked bool `json:"liked"`
}

func updatePostStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)

	log := log.With().Str("endpoint", "/updatePostStatus").Int("userID", userID).Logger()

	var postUpdateData PostUpdate
	err := json.NewDecoder(r.Body).Decode(&postUpdateData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode post from request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log = log.With().Int("postID", postUpdateData.Id).Logger()

	// Invalid state: post can't be liked but not read.
	if postUpdateData.Liked && !postUpdateData.Read {
		log.Error().Msg("Tried to mark post with illegal state")
		http.Error(w, "Can't mark post as liked but not read", http.StatusBadRequest)
		return
	}

	err = updatePostStatus(userID, postUpdateData.Id, postUpdateData.Read, postUpdateData.Liked)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update post status in database")
		writeErrorResponse(err, w)
		return
	}

	// Return the updated post
	post, err := getPost(userID, postUpdateData.Id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get updated post from database")
		writeErrorResponse(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)

	log.Info().Msg("Sucess")
}

func fetchPageHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	userID := getUserIdFromRequest(r)

	log := log.With().Str("endpoint", "/fetchPage").Int("userID", userID).Str("url", url).Logger()

	if url == "" {
		log.Error().Msg("URL missing")
		http.Error(w, "Missing URL", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(url)

	if err != nil {
		log.Warn().Err(err).Msg("Failed to fetch page from URL")
		http.Error(w, fmt.Sprintf("Failed to fetch %s: %v", url, err), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to page fetch read response body")
		http.Error(w, fmt.Sprintf("Failed to read response body: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(bodyBytes)

	log.Info().Msg("Success")
}
func searchPostsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	userID := getUserIdFromRequest(r)

	log := log.With().Str("endpoint", "/searchPosts").Int("userID", userID).Str("query", query).Logger()

	if query == "" {
		log.Warn().Msg("Searching with empty query")
		getPostHandler(w, r)
		return
	}

	searchResultPosts, err := getPostsBySearchInBody(query, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search in posts")
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(searchResultPosts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode posts")
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	log.Info().Msg("Success")
}
