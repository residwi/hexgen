package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"__MODULE__/internal/platform/identity"
	"__MODULE__/internal/platform/web/request"
	"__MODULE__/internal/platform/web/response"
)

func Require(log *slog.Logger, denied string, pred func(identity.Identity) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := request.RequireUser(r)
			if err != nil {
				response.HandleErr(w, err)
				return
			}

			if !pred(id) {
				log.WarnContext(r.Context(), "authorization denied",
					slog.String("reason", denied), slog.String("path", r.URL.Path))
				response.Forbidden(w, denied)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(log *slog.Logger, role string) func(http.Handler) http.Handler {
	return Require(log, fmt.Sprintf("%s access required", role), func(id identity.Identity) bool {
		return id.Role == role
	})
}
