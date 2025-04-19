package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
	"todolist/internal/adapters"
	"todolist/internal/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type TaskBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TaskRequest struct {
	TaskBody
	CategoryIds []uuid.UUID `json:"category_ids"`
}

type OwnershipMiddleware struct {
	userService adapters.UserAdapter
	timeout     time.Duration
}

func NewOwnershipMiddleware(service adapters.UserAdapter, timeout time.Duration) OwnershipMiddleware {
	return OwnershipMiddleware{
		userService: service,
		timeout:     timeout,
	}
}

func (m *OwnershipMiddleware) CheckCategoriesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("CheckCategoriesMiddleware: started processing")

		// saving data for further handlers
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("CheckCategoriesMiddleware: failed to read request body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to read request body"))
			return
		}
		r.Body.Close()

		userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID)
		if !ok {
			log.Warn().
				Msg("CheckCategoriesMiddleware: missing userID in context")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Missing userID"))
			return
		}

		log.Debug().
			Str("userID", userID.String()).
			Msg("CheckCategoriesMiddleware: processing request")

		var req TaskRequest
		err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&req)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("CheckCategoriesMiddleware: failed to decode request body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		log.Debug().
			Int("num_categories", len(req.CategoryIds)).
			Msg("CheckCategoriesMiddleware: checking category ownership")

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, m.timeout)
		defer cancel()

		areCategoriesOwned, err := m.userService.CheckCategoriesOwnership(ctx, userID, req.CategoryIds)
		if err != nil {
			log.Error().
				Err(err).
				Str("userID", userID.String()).
				Msg("CheckCategoriesMiddleware: failed to verify category ownership")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		if !areCategoriesOwned {
			log.Warn().
				Str("userID", userID.String()).
				Int("num_categories", len(req.CategoryIds)).
				Msg("CheckCategoriesMiddleware: unauthorized category access attempt")
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, response.Error("Unauthorized access to categories"))
			return
		}

		log.Info().
			Str("userID", userID.String()).
			Msg("CheckCategoriesMiddleware: successful ownership verification")

		// setting saved data for further use
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		next.ServeHTTP(w, r)
	})
}

func (m *OwnershipMiddleware) CheckTaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("CheckTaskMiddleware: started processing")

		taskID := chi.URLParam(r, "id")
		if taskID == "" {
			log.Warn().
				Msg("CheckTaskMiddleware: empty task ID provided")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("Task ID required"))
			return
		}

		log.Debug().
			Str("taskID", taskID).
			Msg("CheckTaskMiddleware: processing task")

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

		userID, ok := r.Context().Value(UserIDContextKey).(uuid.UUID)
		if !ok {
			log.Warn().
				Msg("CheckTaskMiddleware: missing userID in context")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Missing userID"))
			return
		}

		log.Debug().
			Str("userID", userID.String()).
			Str("taskID", taskUUID.String()).
			Msg("CheckTaskMiddleware: verifying ownership")

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, m.timeout)
		defer cancel()

		isTaskOwned, err := m.userService.CheckTaskOwnership(ctx, userID, taskUUID)
		if err != nil {
			log.Error().
				Err(err).
				Str("userID", userID.String()).
				Str("taskID", taskUUID.String()).
				Msg("CheckTaskMiddleware: failed to verify task ownership")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		if !isTaskOwned {
			log.Warn().
				Str("userID", userID.String()).
				Str("taskID", taskUUID.String()).
				Msg("CheckTaskMiddleware: unauthorized task access attempt")
			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, response.Error("Unauthorized access to task"))
			return
		}

		log.Info().
			Str("userID", userID.String()).
			Str("taskID", taskUUID.String()).
			Msg("CheckTaskMiddleware: ownership verified successfully")

		next.ServeHTTP(w, r)
	})
}
