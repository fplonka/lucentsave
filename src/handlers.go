package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx"
	"golang.org/x/crypto/bcrypt"
)

func addHandleFuncs() {
	// static stuff
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeCacheHeader(maxCacheTimeout, w)
		fs.ServeHTTP(w, r)
	})))
	http.Handle("/favicon.ico", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeCacheHeader(maxCacheTimeout, w)
		http.ServeFile(w, r, "static/favicon.ico")
	}))

	// POST
	http.HandleFunc("/mark-liked", authMiddleware(markLikedHandler))
	http.HandleFunc("/mark-read", authMiddleware(markReadHandler))
	http.HandleFunc("/update-post-state", authMiddleware(updatePostStateHandler))
	http.HandleFunc("/save", authMiddleware(savePostHandler))
	http.HandleFunc("/delete-post", authMiddleware(deletePostHandler))
	http.HandleFunc("/create-user", createUserHandler) // registration attempt

	// GET
	http.HandleFunc("/post", authMiddleware(postStaticHandler))
	http.HandleFunc("/post-status", authMiddleware(postStatusHandler))
	http.HandleFunc("/fetch-url", authMiddleware(fetchURL))
	http.HandleFunc("/saved", authMiddleware(getPostListHandler("/saved")))
	http.HandleFunc("/read", authMiddleware(getPostListHandler("/read")))
	http.HandleFunc("/search", authMiddleware(getPostListHandler("/search")))
	http.HandleFunc("/query", authMiddleware(queryHandler))
	http.HandleFunc("/", redirectIfSignedInMiddelware(signinPageHandler))           // sign in page
	http.HandleFunc("/signin", redirectIfSignedInMiddelware(signinPageHandler))     // sign in page
	http.HandleFunc("/register", redirectIfSignedInMiddelware(registerPageHandler)) // registration page
	http.HandleFunc("/authenticate", authenticateHandler)                           // sign in attempt
}

func writeCacheHeader(duration int, w http.ResponseWriter) {
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%v", duration))
}

const maxCacheTimeout = 365 * 24 * 60 * 60

func getUserIdFromRequest(r *http.Request) int {
	return r.Context().Value(userIDKey).(int)
}

func respondInternalError(w http.ResponseWriter) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func respondBadRequest(w http.ResponseWriter) {
	http.Error(w, "bad request", http.StatusBadRequest)
}

func respondUnauthorized(w http.ResponseWriter) {
	http.Error(w, "unauthorized", http.StatusUnauthorized)

}

func logAndRespondInternalError(logger *slog.Logger, msg string, w http.ResponseWriter, err error, attr ...any) {
	args := append([]any{"error", err}, attr...)
	logger.Error(msg, args...)
	respondInternalError(w)
}

func getPostListHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getUserIdFromRequest(r)

		logger := slog.Default().With("func", "getPostListHandler", "path", path, "userID", userID)

		var postEntries []Post
		data := map[string]any{}
		data["Path"] = path

		switch path {
		case "/saved":
			postEntries = getUserPostsInfo(userID, false)
			data["Saved"] = true
		case "/read":
			postEntries = getUserPostsInfo(userID, true)
			data["Read"] = true
		case "/search":
			postEntries = []Post{}
			data["Search"] = true
		}

		data["Posts"] = postEntries

		w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0")
		var err error
		if r.Header.Get("HX-Request") == "true" {
			logAndRespondInternalError(logger, "hx-request in path?!", w, nil)
			return
		} else {
			err = postListTemplate.ExecuteTemplate(w, "base", data)
		}

		if err != nil {
			logAndRespondInternalError(logger, "post list template error", w, nil)
			return
		}
	}
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	userID := getUserIdFromRequest(r)

	if !r.Form.Has("query") {
		respondBadRequest(w)
		return
	}

	query := r.Form.Get("query")

	if query == "" {
		respondBadRequest(w)
		return
	}

	logger := slog.Default().With("func", "queryHandler", "userID", userID, "query", query)

	// postEntries := searchUserPosts(userID, query)
	querryEmbedding, err := getEmbedding(query)
	if err != nil {
		logAndRespondInternalError(logger, "failed to get query embedding", w, err)
		return
	}
	postEntries := searchUserPostsByEmbedding(userID, querryEmbedding)

	// TODO: this might be a good spot to cache with etags, search is expensive..
	err = postListTemplate.ExecuteTemplate(w, "postList", map[string][]Post{"Posts": postEntries})
	if err != nil {
		logAndRespondInternalError(logger, "failed to get execute search result postList template", w, err)
		return
	}

}

func markLikedHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	postID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		respondBadRequest(w)
		return
	}

	isLiked := false
	if r.Form.Has("liked") {
		isLiked = true
	}

	err = markPostLiked(postID, isLiked)
	if err != nil {
		respondInternalError(w)
		return
	}
}

func markReadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	postID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		respondBadRequest(w)
		return
	}

	isRead := false
	if r.Form.Has("read") {
		isRead = true
	}

	err = markPostRead(postID, isRead)
	if err != nil {
		respondInternalError(w)
	}
}

func updatePostStateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	postID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		respondBadRequest(w)
		return
	}

	userID := getUserIdFromRequest(r)

	isRead := r.FormValue("read") != ""
	isLiked := r.FormValue("liked") != "" && isRead // Ensure isLiked is only true if isRead is also true

	err = updatePostStatus(postID, userID, isRead, isLiked)
	if err != nil {
		respondInternalError(w)
		return
	}
}

func savePostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.Form.Get("title")
	url := r.Form.Get("url")
	content := r.Form.Get("content")

	userID := getUserIdFromRequest(r)

	logger := slog.Default().With("func", "savePostHandler", "userID", userID)

	totalLength := len(title) + len(url) + len(content)
	if totalLength > 200000 {
		logger.Warn("post too long", "url", url, "length", totalLength)
		respondBadRequest(w)
		return
	}

	post := Post{URL: url, Title: title, Body: content, TimeAdded: time.Now().Unix(), UserID: userID}
	postID, err := savePost(post)
	if err != nil {
		respondInternalError(w)
		return
	}

	post.ID = postID

	go saveEmbedding(post)

	// we need to pass in numbers for Index and Last so that the post doesn't think it's the latpost, since then it wouldn't render the separator
	err = postListTemplate.ExecuteTemplate(w, "postEntry", map[string]any{"Post": post, "Index": 0, "Total": 0})
	if err != nil {
		logger = logger.With("postID", postID)
		logAndRespondInternalError(logger, "failed to execute template", w, err)
		return
	}
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	postID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		respondBadRequest(w)
	}

	userID := getUserIdFromRequest(r)

	err = deletePost(userID, postID)
	if err != nil {
		respondInternalError(w)
		return
	}

	w.Header().Set("HX-Redirect", "/saved")
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// TODO this is so shit jakdfljkldafjkldjf
func fetchURL(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	url := r.URL.Query().Get("url")

	if !isUrl(url) {
		respondBadRequest(w)
		return
	}

	logger := slog.Default().With("func", "fetchURL", "url", url)

	// Fetch the URL server-side, check for errors
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		logger.Warn("failed to fetch url")
		respondBadRequest(w)
		return
	}
	// TODO: check for 4xx??

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logAndRespondInternalError(logger, "failed to read response body", w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(body)
}

func postStaticHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	userID := getUserIdFromRequest(r)
	postID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		respondBadRequest(w)
		return
	}
	post, err := getPostContent(postID, userID)
	if err != nil {
		respondInternalError(w)
		return
	}

	logger := slog.Default().With("func", "postStaticHandler", "userID", userID, "postID", postID)

	writeCacheHeader(30*24*60*60, w) // month
	w.Header().Set("Vary", "HX-Request")
	if r.Header.Get("HX-Request") == "true" {
		logAndRespondInternalError(logger, "shouldnt ever happen?!?!", w, err)
		return
	} else {
		err = postViewTemplate.ExecuteTemplate(w, "base", map[string]any{"Post": post})
	}
	if err != nil {
		logAndRespondInternalError(logger, "failed to execute post view template", w, err)
		return
	}
}

func postStatusHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	userID := getUserIdFromRequest(r)
	postID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		respondBadRequest(w)
		return
	}
	post, err := getPostContent(postID, userID)
	if err != nil {
		respondInternalError(w)
		return
	}

	logger := slog.Default().With("func", "postStatusHandler", "userID", userID, "postID", postID)

	w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0")
	if r.Header.Get("HX-Request") == "true" {
		err = postViewTemplate.ExecuteTemplate(w, "postStatus", map[string]any{"Post": post})
	} else {
		logAndRespondInternalError(logger, "shouldnt ever happen?!?!", w, err)
		return
	}
	if err != nil {
		logAndRespondInternalError(logger, "failed to execute post view template", w, err)
	}
}

func signinPageHandler(w http.ResponseWriter, r *http.Request) {
	writeCacheHeader(maxCacheTimeout, w)
	w.Header().Set("Vary", "HX-Request")

	var err error
	if r.Header.Get("HX-Request") == "true" {
		err = signinTemplate.ExecuteTemplate(w, "signInForm", nil)
	} else {
		err = signinTemplate.ExecuteTemplate(w, "base", map[string]any{"isSignIn": true})
	}

	if err != nil {
		logger := slog.Default().With("func", "signinPageHandler")
		logAndRespondInternalError(logger, "failed to execute sign in page template", w, err)
	}
}

func registerPageHandler(w http.ResponseWriter, r *http.Request) {
	writeCacheHeader(maxCacheTimeout, w)
	w.Header().Set("Vary", "HX-Request")

	var err error
	if r.Header.Get("HX-Request") == "true" {
		err = signinTemplate.ExecuteTemplate(w, "registerForm", nil)
	} else {
		err = signinTemplate.ExecuteTemplate(w, "base", map[string]any{"isSignIn": false})
	}

	if err != nil {
		logger := slog.Default().With("func", "registerPageHandler")
		logAndRespondInternalError(logger, "failed to execute register page template", w, err)
	}
}

// TODO: error messages on frontend
func authenticateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	email := r.Form.Get("email")
	if email == "" {
		respondBadRequest(w)
		return
	}

	hashedPassword, userID, err := getHashedPasswordAndUserId(email)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "no user found with provided email", http.StatusUnauthorized)
		} else {
			http.Error(w, "authentication failed", http.StatusUnauthorized)
		}
		return
	}

	password := r.Form.Get("password")
	if password == "" {
		http.Error(w, "password not provided", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	// w.Header().Set("HX-Redirect", "/saved")
	generateAndSetAuthToken(w, userID)
	http.Redirect(w, r, "/saved", http.StatusTemporaryRedirect)
}

// TODO: error messages on frontend
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	email := r.Form.Get("email")
	if email == "" {
		http.Error(w, "email not provided", http.StatusBadRequest)
		return
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}

	taken, err := checkUserExists(email)
	if err != nil {
		http.Error(w, "failed to create account", http.StatusInternalServerError)
		return
	}

	if taken {
		http.Error(w, "email already in use", http.StatusBadRequest)
		return
	}

	password := r.Form.Get("password")
	if password == "" {
		http.Error(w, "password not provided", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		http.Error(w, "failed to create account", http.StatusInternalServerError)
		return
	}

	id, err := createUser(email, string(hashedPassword))

	if err != nil {
		http.Error(w, "failed to create account", http.StatusInternalServerError)
		return
	}

	generateAndSetAuthToken(w, id)
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			slog.Error("failed to parse request form", "error", err, "form", r.Form)
			return
		}

		startTime := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(startTime)

		slog.Info("req", "method", r.Method, "path", r.URL.Path, "duration", duration.Milliseconds())
	})
}
