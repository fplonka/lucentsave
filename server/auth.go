package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/rs/zerolog/log"
)

type Claims struct {
	UserID int `json:"userID"`
	jwt.StandardClaims
}

func generateAndSetAuthToken(w http.ResponseWriter, userID int) error {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 1 week
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:         userID,
		StandardClaims: jwt.StandardClaims{ExpiresAt: expirationTime.Unix()},
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return err
	}

	// Set the new token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: DEFAULT_SAME_SITE_MODE,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "loggedIn",
		Value:    "true",
		Expires:  expirationTime,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: DEFAULT_SAME_SITE_MODE,
		Path:     "/",
	})

	return nil
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
				log.Warn().Err(err).Msg("No token cookie provided")
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusUnauthorized)
			log.Error().Err(err).Msg("Failed to get request's token cookie")
			return
		}

		tokenString := c.Value
		claims := &Claims{}

		tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				log.Warn().Err(err).Msg("Invalid token signature")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			log.Error().Err(err).Msg("Failed to parse token string")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !tkn.Valid {
			log.Warn().Err(err).Msg("Invalid auth token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// JWT is valid so we refresh it
		err = generateAndSetAuthToken(w, claims.UserID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate and set auth token")
			w.WriteHeader(http.StatusUnauthorized)
		}

		// We add the user ID to the context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		// Call the next handler function with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
