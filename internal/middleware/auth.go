package middleware

import (
	"context"
	"net/http"
	"strings"
	auth_utils "todolist/internal/pkg/authUtils"
	"todolist/internal/pkg/response"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

var (
	UserIDContextKey string = "contextKeyID{}"
)

func NewJwtAuthMiddleware(secretSrc string, tokenHandlerSrc auth_utils.ITokenHandler) JwtAuthMiddleware {
	return JwtAuthMiddleware{
		secret:       secretSrc,
		tokenHandler: tokenHandlerSrc,
	}
}

type JwtAuthMiddleware struct {
	secret       string
	tokenHandler auth_utils.ITokenHandler
}

func (m *JwtAuthMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			log.Info().Msg("user with no token came")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("Error in parsing token"))
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")

		payload, err := m.tokenHandler.ParseToken(token, m.secret)
		if err != nil {
			if err == auth_utils.ErrParsingToken {
				log.Info().Msg("user with invalid jwt came")
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, response.Error(err.Error()))
			} else {
				log.Info().Msg("user with invalid jwt came")
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.Error(err.Error()))
			}
			return
		}
		ctx := context.WithValue(r.Context(), UserIDContextKey, payload.ID)

		log.Info().Msgf("user with id %v successfully authorized", payload.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
