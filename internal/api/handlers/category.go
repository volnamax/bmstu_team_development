package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"
	"todolist/internal/middleware"
	"todolist/internal/models"
	"todolist/internal/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type CategoryBody struct {
	Name string `json:"name"`
}

type CategoryResponse struct {
	ID uuid.UUID `json:"id"`
	CategoryBody
}

type CategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

type CategoriesProvider interface {
	CreateCategory(ctx context.Context, category *models.CategoryBody) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context, pageIndex, recordsPerPage int, userid uuid.UUID) ([]models.Category, error)
}

// @Summary CreateCategory
// @Security ApiKeyAuth
// @Tags category
// @Description create category
// @ID create-category
// @Accept  json
// @Produce  json
// @Param input body CategoryBody true "category name"
// @Success 200
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/category [post]
func CreateCategory(categoryProvider CategoriesProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
		if !ok {
			log.Warn().
				Msg("CreateCategory: missing userID")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Missing userID"))
			return
		}

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("CreateCategory: started processing request")

		var req CategoryBody
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("CreateCategory: failed to decode request body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}
		category := models.CategoryBody{Name: req.Name, UserID: userID}

		log.Debug().
			Str("category_name", req.Name).
			Msg("CreateCategory: received create request")

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err = categoryProvider.CreateCategory(ctx, &category)
		if err != nil {
			log.Error().
				Err(err).
				Str("category_name", req.Name).
				Msg("CreateCategory: failed to create category")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		log.Info().
			Str("category_name", req.Name).
			Msg("CreateCategory: successfully created new category")
		render.Status(r, http.StatusOK)
	}
}

// @Summary DeleteCategory
// @Security ApiKeyAuth
// @Tags category
// @Description delete category
// @ID delete-category
// @Accept  json
// @Produce  json
// @Param id   path      string  true  "Category ID (UUID)"
// @Success 200
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/category/{id} [delete]
func DeleteCategory(categoryProvider CategoriesProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("DeleteCategory: started processing request")

		id := chi.URLParam(r, "id")
		log.Debug().
			Str("category_id", id).
			Msg("DeleteCategory: received category ID")

		uuid, err := uuid.Parse(id)
		if err != nil {
			log.Warn().
				Str("category_id", id).
				Err(err).
				Msg("DeleteCategory: invalid category UUID format")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid UUID"))
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err = categoryProvider.Delete(ctx, uuid)
		if err != nil {
			if errors.Is(err, models.ErrCategoryNotFound) {
				log.Warn().
					Str("category_id", uuid.String()).
					Msg("DeleteCategory: category not found")
				render.Status(r, http.StatusNotFound)
			} else {
				log.Error().
					Str("category_id", uuid.String()).
					Err(err).
					Msg("DeleteCategory: failed to delete category")
				render.Status(r, http.StatusInternalServerError)
			}
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		log.Info().
			Str("category_id", uuid.String()).
			Msg("DeleteCategory: successfully deleted category")
		render.Status(r, http.StatusOK)
	}
}

// @Summary GetCategories
// @Security ApiKeyAuth
// @Tags category
// @Description get all categories
// @ID get-categories
// @Accept  json
// @Produce  json
// @Param input body Pagination true "pagination info"
// @Success 200 {object} CategoriesResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/category/all [post]
func GetCategories(categoryProvider CategoriesProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("GetCategories: started processing request")

		userID, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
		if !ok {
			log.Warn().
				Msg("GetCategories: failed to get UserID")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("Missing userID"))
			return
		}

		var req Pagination
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("GetCategories: failed to decode request body")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		log.Debug().
			Int("page_index", req.PageIndex).
			Int("records_per_page", req.RecordsPerPage).
			Msg("GetCategories: received pagination parameters")

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		var categories []models.Category
		categories, err = categoryProvider.GetAll(ctx, req.PageIndex, req.RecordsPerPage, userID)
		if err != nil {
			log.Error().
				Err(err).
				Int("page_index", req.PageIndex).
				Int("records_per_page", req.RecordsPerPage).
				Msg("GetCategories: failed to fetch categories")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		log.Info().
			Int("count", len(categories)).
			Msg("GetCategories: successfully fetched categories")

		render.Status(r, http.StatusOK)
		render.JSON(w, r, toCategoriesResponse(categories))
	}
}

func toCategoryResponse(category models.Category) CategoryResponse {
	return CategoryResponse{
		ID: category.ID,
		CategoryBody: CategoryBody{
			Name: category.Name,
		},
	}
}

func toCategoriesResponse(categories []models.Category) CategoriesResponse {
	responses := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		responses = append(responses, toCategoryResponse(category))
	}
	return CategoriesResponse{
		Categories: responses,
	}
}
