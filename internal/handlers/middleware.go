package handlers

import (
	"context"
	"net/http"
)

type userKeyContext string

func (handler Handler) AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		user, err := handler.AuthService.GetUserByToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userKeyContext("user"), *user)
		authR := r.WithContext(ctx)
		h.ServeHTTP(w, authR)
	})
}
