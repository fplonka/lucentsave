package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type UserClaims struct {
	UserID int `json:"userID"`
	jwt.StandardClaims
}

func generateAndSetAuthToken(w http.ResponseWriter, userID int) error {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 1 week

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		userID,
		jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	return nil
}

func getRequestToken(r *http.Request) (*jwt.Token, error) {
	c, err := r.Cookie("token")
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(c.Value, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token invalid")
	}

	return token, nil
}

type key int

const (
	userIDKey key = iota
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := getRequestToken(r)
		if err != nil {
			// TODO: bug, maybe fix one day. login in, delete auth cookie, try to save post.
			// unauthed so we get redirect to /signin, but hx-request is true after the
			// redirect and the sign in form fragment gets added to the post list by htmx...
			slog.Warn("invalid token or no token")
			http.Redirect(w, r, "/signin", http.StatusTemporaryRedirect)
			return
		}

		// Set user ID (extracted from token) in context
		claims, ok := token.Claims.(*UserClaims)
		if !ok {
			logAndRespondInternalError(slog.Default(), "failed to extract token claims", w, err)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), userIDKey, claims.UserID))
		next.ServeHTTP(w, r)
	})
}

func redirectIfSignedInMiddelware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := getRequestToken(r)
		if err == nil {
			http.Redirect(w, r, "/saved", http.StatusTemporaryRedirect)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
