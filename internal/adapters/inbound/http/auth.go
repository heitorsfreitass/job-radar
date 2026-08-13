package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 7 * 24 * time.Hour

type contextKey string

const userIDContextKey contextKey = "userID"

func issueToken(secret []byte, userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func parseToken(secret []byte, tokenString string) (userID string, err error) {
	claims := jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}

// RequireAuth builds middleware that rejects requests without a valid
// `Authorization: Bearer <token>` header, and stashes the authenticated
// user's id in the request context (see userIDFromContext).
func RequireAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				writeError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}

			userID, err := parseToken(secret, token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(userIDContextKey).(string)
	return id
}
