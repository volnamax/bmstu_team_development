package middleware

import (
	"context"
	"net/http"
	"strings"
	"todolist/internal/api/handlers"
	auth_utils "todolist/internal/pkg/authUtils"
	"todolist/internal/pkg/response"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
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
			render.JSON(w, r, response.Error("Error in parsing token"))
			render.Status(r, http.StatusBadRequest)
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")

		payload, err := m.tokenHandler.ParseToken(token, m.secret)
		if err != nil {
			if err == auth_utils.ErrParsingToken {
				log.Info().Msg("user with invalid jwt came")
				render.JSON(w, r, response.Error(err.Error()))
				render.Status(r, http.StatusBadRequest)
			} else {
				log.Info().Msg("user with invalid jwt came")
				render.JSON(w, r, response.Error(err.Error()))
				render.Status(r, http.StatusUnauthorized)
			}
			return
		}
		ctx := context.WithValue(r.Context(), handlers.UserIDContextKey, payload.ID)

		log.Info().Msgf("user with id %v successfully authorized", payload.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
