package apiserver

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"asyncapi/store"

	"github.com/google/uuid"
)

type userCtxKey struct {
}

func ContextWithUser(ctx context.Context, user *store.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func NewLoggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Request received",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("Actual Port", r.RequestURI),
			)
			next.ServeHTTP(w, r)
			logger.Info("Response sent",
				// Assuming you have a way to get the response status code
				//slog.String("response_body", "response body"),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

func NewAuthMiddleware(jwtManager *JwtManager, userStore *store.UserStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/auth") {
				next.ServeHTTP(w, r)
				return
			}
			//Authorisation header
			//read auth header
			authHeader := r.Header.Get("Authorization")
			var token string
			if parts := strings.Split(authHeader, "Bearer "); len(parts) == 2 {
				token = parts[1]
			}
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			parsedToken, err := jwtManager.Parse(token)
			if err != nil {
				slog.Error("failed to parse the token", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			//verify the parsed token
			if !jwtManager.IsAccessToken(parsedToken) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("not an access token"))
				return
			}

			//userId from claims
			userIdStr, err := parsedToken.Claims.GetSubject()
			if err != nil {
				slog.Error("faield to extract user info from claims subject", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			//convert userId as uuid
			userId, err := uuid.Parse(userIdStr)
			if err != nil {
				slog.Error("failed to parse the userId into uuid type", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			//check if user exists in db
			//will try to move this to cache, later
			user, err := userStore.GetUserByID(r.Context(), userId)
			if err != nil {
				slog.Error("user could not be found in database", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), user)))

		})
	}
}
