package middleware

import (
	"context"
	"net/http"
	"time"
	"todolist/internal/adapters"
	"todolist/internal/api/handlers"
	"todolist/internal/pkg/response"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type OwnershipMiddleware struct {
	userService adapters.UserAdapter
	timeout     time.Duration
}

func NewOwnershipMiddleware(service adapters.UserAdapter, timeout time.Duration) OwnershipMiddleware {
	return OwnershipMiddleware{
		userService: service,
	}
}

func (m *OwnershipMiddleware) CheckCategoriesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("CheckCategoriesMiddleware: started request processing")

		userID, ok := r.Context().Value(handlers.UserIDContextKey).(string)
		if !ok {
			log.Warn().
				Str("reason", "missing_user_id").
				Msg("CheckCategoriesMiddleware: missing/invalid userID in context")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Missing userID"))
			return
		}

		log.Debug().Str("userID", userID).Msg("CheckCategoriesMiddleware: processing user")

		userUUID, err := uuid.Parse(userID)
		if err != nil {
			log.Warn().
				Str("userID", userID).
				Err(err).
				Msg("CheckCategoriesMiddleware: malformed userID UUID")
			render.JSON(w, r, response.Error("Invalid userID"))
			render.Status(r, http.StatusBadRequest)
			return
		}

		var req handlers.TaskRequest
		err = render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("CheckCategoriesMiddleware: failed to decode request body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		log.Debug().
			Interface("category_ids", req.CategoryIds).
			Msg("CheckCategoriesMiddleware: checking ownership for categories")

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, m.timeout)
		defer cancel()

		areCategoriesOwned, err := m.userService.CheckCategoriesOwnership(ctx, userUUID, req.CategoryIds)
		if err != nil {
			log.Error().
				Err(err).
				Str("userID", userUUID.String()).
				Interface("category_ids", req.CategoryIds).
				Msg("CheckCategoriesMiddleware: failed to check categories ownership")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		if !areCategoriesOwned {
			log.Warn().
				Str("userID", userUUID.String()).
				Interface("category_ids", req.CategoryIds).
				Msg("CheckCategoriesMiddleware: user doesn't own one or more categories")
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, response.Error("Unauthorized access to categories"))
			return
		}

		log.Info().
			Str("userID", userUUID.String()).
			Msg("CheckCategoriesMiddleware: successful ownership verification")
		next.ServeHTTP(w, r)
	})
}

func (m *OwnershipMiddleware) CheckTaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("CheckTaskMiddleware: started request processing")

		taskID := chi.URLParam(r, "id")
		log.Debug().Str("taskID", taskID).Msg("CheckTaskMiddleware: processing task")

		taskUUID, err := uuid.Parse(taskID)
		if err != nil {
			log.Warn().
				Str("taskID", taskID).
				Err(err).
				Msg("CheckTaskMiddleware: invalid task UUID format")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid UUID"))
			return
		}

		userID, ok := r.Context().Value(handlers.UserIDContextKey).(string)
		if !ok {
			log.Warn().
				Str("reason", "missing_user_id").
				Msg("CheckTaskMiddleware: missing/invalid userID in context")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Missing userID"))
			return
		}

		log.Debug().Str("userID", userID).Msg("CheckTaskMiddleware: processing user")

		userUUID, err := uuid.Parse(userID)
		if err != nil {
			log.Warn().
				Str("userID", userID).
				Err(err).
				Msg("CheckTaskMiddleware: malformed userID UUID")
			render.JSON(w, r, response.Error("Invalid userID"))
			render.Status(r, http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, m.timeout)
		defer cancel()

		isTaskOwned, err := m.userService.CheckTaskOwnership(ctx, userUUID, taskUUID)
		if err != nil {
			log.Error().
				Err(err).
				Str("userID", userUUID.String()).
				Str("taskID", taskUUID.String()).
				Msg("CheckTaskMiddleware: failed to check task ownership")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		if !isTaskOwned {
			log.Warn().
				Str("userID", userUUID.String()).
				Str("taskID", taskUUID.String()).
				Msg("CheckTaskMiddleware: user doesn't own the task")
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, response.Error("Unauthorized access to task"))
			return
		}

		log.Info().
			Str("userID", userUUID.String()).
			Str("taskID", taskUUID.String()).
			Msg("CheckTaskMiddleware: successful ownership verification")
		next.ServeHTTP(w, r)
	})
}
