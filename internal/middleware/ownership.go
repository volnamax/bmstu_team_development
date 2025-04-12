package middleware

import (
	"context"
	"net/http"
	"time"
	"todolist/internal/adapters"
	"todolist/internal/pkg/response"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
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
	}
}

func (m *OwnershipMiddleware) CheckCategoriesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDContextKey).(string)
		if !ok {
			//m.logger.Warn("Missing/invalid userID")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Missing userID"))
			return
		}

		userUUID, err := uuid.Parse(userID)
		if err != nil {
			//m.logger.WithError(err).
			//	WithField("userID", userID).
			//	Warn("Malformed userID UUID")
			render.JSON(w, r, response.Error("Invalud userID"))
			render.Status(r, http.StatusBadRequest)
			return
		}

		var req TaskRequest
		err = render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, m.timeout)
		defer cancel()

		areCategoriesOwned, err := m.userService.CheckCategoriesOwnership(ctx, userUUID, req.CategoryIds)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
		}
		if !areCategoriesOwned {
			render.Status(r, http.StatusForbidden)
		}

		next.ServeHTTP(w, r)
	})
}

func (m *OwnershipMiddleware) CheckTaskMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskID := chi.URLParam(r, "id")

		taskUUID, err := uuid.Parse(taskID)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid UUID"))
			return
		}

		userID, ok := r.Context().Value(UserIDContextKey).(string)
		if !ok {
			//m.logger.Warn("Missing/invalid userID")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Missing userID"))
			return
		}

		userUUID, err := uuid.Parse(userID)
		if err != nil {
			//m.logger.WithError(err).
			//	WithField("userID", userID).
			//	Warn("Malformed userID UUID")
			render.JSON(w, r, response.Error("Invalud userID"))
			render.Status(r, http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, m.timeout)
		defer cancel()

		isTaskOwned, err := m.userService.CheckTaskOwnership(ctx, userUUID, taskUUID)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
		}
		if !isTaskOwned {
			render.Status(r, http.StatusForbidden)
		}

		next.ServeHTTP(w, r)

	})
}
