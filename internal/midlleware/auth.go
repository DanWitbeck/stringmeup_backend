// internal/middleware/auth.go
package midlleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/threadcraft-backend/internal/db"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				db.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token")
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims,
				func(t *jwt.Token) (any, error) {
					return []byte(jwtSecret), nil
				},
				jwt.WithValidMethods([]string{"HS256"}),
			)
			if err != nil || !token.Valid {
				db.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(r *http.Request) string {
	id, _ := r.Context().Value(UserIDKey).(string)
	return id
}
