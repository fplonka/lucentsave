package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"userID"`
	jwt.StandardClaims
}

func generateAndSetAuthToken(w http.ResponseWriter, userID int, username string) {
	expirationTime := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:         userID,
		Username:       username,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Failed to authorize user", http.StatusInternalServerError)
		return
	}

	// Set the new token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "loggedIn",
		Value:   "true",
		Expires: expirationTime,
		Path:    "/",
	})
}

func authorizeAndWriteToken(w http.ResponseWriter, user User) {
	var userID int
	var hashedPassword []byte
	err := db.QueryRow("SELECT id, password FROM users WHERE username = $1", user.Username).Scan(&userID, &hashedPassword)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(user.HashedPassword)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	generateAndSetAuthToken(w, userID, user.Username)

	// Respond to the client
	w.WriteHeader(http.StatusCreated)
}

type key int

const (
	userIDKey key = iota
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenString := c.Value
		claims := &Claims{}

		tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// JWT is valid so we refresh it
		generateAndSetAuthToken(w, claims.UserID, claims.Username)

		// We add the user ID to the context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		// Call the next handler function with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func trace(s string) (string, time.Time) {
	log.Println("START:", s)
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("  END:", s, "ElapsedTime in seconds:", endTime.Sub(startTime))
}
