package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

func addHandleFuncs(mux *http.ServeMux) {
	// TODO: middlewhere for checking post owner id == user id ??
	mux.HandleFunc("/api/getAllUserPosts", authMiddleware(getSavedPostsHandler))
	mux.HandleFunc("/api/getPost", authMiddleware(getPostHandler))
	mux.HandleFunc("/api/createPost", authMiddleware(createPostHandler))
	mux.HandleFunc("/api/deletePost", authMiddleware(deletePostHandler))
	mux.HandleFunc("/api/updatePostStatus", authMiddleware(updatePostStatusHandler))
	mux.HandleFunc("/api/createUser", createUserHandler)
	mux.HandleFunc("/api/signout", signoutHandler)
	mux.HandleFunc("/api/signin", signinHandler)
	mux.HandleFunc("/api/fetchPage", fetchPageHandler)
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

func writePostsListResponse(userID int, w http.ResponseWriter) {
	posts, err := getUserPosts(userID)
	if err != nil {
		writeErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func getUserIdFromRequest(r *http.Request) int {
	return r.Context().Value(userIDKey).(int)
}

// Gets all posts for a given user, regardless of whether they're read or liked
func getSavedPostsHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)
	writePostsListResponse(userID, w)
}

func getPostHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)
	queryParams := r.URL.Query()
	postIDStr := queryParams.Get("id")
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Post ID must be an integer", http.StatusBadRequest)
		return
	}

	post, err := getPost(userID, postID)
	if err != nil {
		writeErrorResponse(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)

	var post Post

	// Decode the incoming Post json
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		writeErrorResponse(err, w)
		return
	}

	err = addPost(post, userID)
	if err != nil {
		writeErrorResponse(err, w)
		return
	}

	// Return update posts list
	writePostsListResponse(userID, w)
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)
	queryParams := r.URL.Query()
	postIDStr := queryParams.Get("id")
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Post ID must be an integer", http.StatusBadRequest)
		return
	}

	err = deletePost(userID, postID)
	if err != nil {
		writeErrorResponse(err, w)
		return
	}

	// Get the new list of posts and return that as JSON.
	// TODO: extract this out
	writePostsListResponse(userID, w)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User

	// Decode the incoming User json
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if email is valid
	_, err = mail.ParseAddress(user.Email)
	if err != nil {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	// Check if the user already exists
	emailTaken := checkEmailIsUsed(user.Email)
	if emailTaken {
		http.Error(w, "Email already taken", http.StatusBadRequest)
		return
	}

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	// Insert the user into the database.
	_, err = addUser.Exec(user.Email, hashedPassword)
	if err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	authorizeAndWriteToken(w, user)
}

func signinHandler(w http.ResponseWriter, r *http.Request) {

	var user User

	// Decode the incoming User json
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !checkEmailIsUsed(user.Email) {
		http.Error(w, "Account not found", http.StatusBadRequest)
		return
	}

	authorizeAndWriteToken(w, user)
}

func signoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		MaxAge:   -1, // Delete cookie immediately
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "loggedIn",
		Value:  "",
		MaxAge: -1,
		Secure: true,
		Path:   "/",
	})
	w.WriteHeader(http.StatusOK)
}

type PostUpdate struct {
	Id    int  `json:"id"`
	Read  bool `json:"read"`
	Liked bool `json:"liked"`
}

func updatePostStatusHandler(w http.ResponseWriter, r *http.Request) {
	userID := getUserIdFromRequest(r)

	var postData PostUpdate
	err := json.NewDecoder(r.Body).Decode(&postData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Invalid state: post can't be liked but not read.
	if postData.Liked && !postData.Read {
		http.Error(w, "Can't mark post as liked but not read", http.StatusBadRequest)
		return
	}

	err = updatePostStatus(userID, postData.Id, postData.Read, postData.Liked)
	if err != nil {
		writeErrorResponse(err, w)
		return
	}

	// Return the update post
	post, err := getPost(userID, postData.Id)
	if err != nil {
		writeErrorResponse(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func fetchPageHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")

	if url == "" {
		http.Error(w, "Missing URL", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(url)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch %s: %v", url, err), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response body: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(bodyBytes)
}
