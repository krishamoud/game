// Package middleware is reponsible for all api middlewares
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type ctxKey int

const (
	ctxUserID ctxKey = iota
)

// Authenticate is a middleware that reads the Authorization header to find a
// jwt Token and rejects requests without a token.  The token is then parsed and
// verified agains the "secret" password.
// TODO: change "secret" to something secret
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		// Get token from the Authorization header
		// format: Authorization: Bearer
		tokens, ok := r.Header["Authorization"]
		if ok && len(tokens) >= 1 {
			token = tokens[0]
			token = strings.TrimPrefix(token, "Bearer ")
		}

		// If the token is empty...
		if token == "" {
			// If we get here, the required token is missing
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Now parse the token
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				msg := fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				return nil, msg
			}
			return []byte("secret"), nil
		})
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check token is valid
		if parsedToken != nil && parsedToken.Valid {
			userID := parsedToken.Claims.(jwt.MapClaims)["_id"]
			if userID == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Everything worked! Set the user in the context.
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Token is invalid
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	})
}

// UserID returns the UserID from the context of authenticated requests
func UserID(c context.Context) string {
	return c.Value(ctxUserID).(string)
}
