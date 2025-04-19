package handlers

import (
	"context"
	"net/http"
	"time"
	"todolist/internal/middleware"
	"todolist/internal/models"
	"todolist/internal/pkg/response"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type TaskBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TaskMeta struct {
	ID     uuid.UUID `json:"id"`
	IsDone bool      `json:"is_done"`
}

type TaskRequest struct {
	TaskBody
	CategoryIds []uuid.UUID `json:"category_ids"`
}

type TaskResponse struct {
	TaskMeta
	TaskBody
	CategoriesResponse
}

type TaskShortResponse struct {
	TaskMeta
	Title string `json:"title"`
}

type TasksList struct {
	List []TaskShortResponse `json:"list"`
}

type TaskProvider interface {
	CreateTask(ctx context.Context, userId uuid.UUID, body *models.TaskBody, categoryIDs []uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, body *models.TaskBody, categoryIDs []uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.TaskFullInfo, error)
	GetAll(ctx context.Context, userId uuid.UUID, pageIndex, recordsPerPage int) ([]models.TaskShortInfo, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ToggleDone(ctx context.Context, id uuid.UUID) error
}

// @Summary CreateTask
// @Security ApiKeyAuth
// @Tags task
// @Description Создать новую задачу
// @ID create-task
// @Accept  json
// @Produce  json
// @Param input body TaskRequest true "task info"
// @Success 200
// @Failure 400,401 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task [post]
func CreateTask(taskProvider TaskProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Trace().Msg("get CreateTask request")

		var req TaskRequest
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse request")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		userId, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
		if !ok {
			log.Error().Msg("no uuid in context")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("unauthorized"))
			return
		}

		err = taskProvider.CreateTask(ctx, userId, toModelTaskBody(req), req.CategoryIds)
		if err != nil {
			log.Err(err).Msg("CreateTask, error from provider")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		render.Status(r, http.StatusOK)
	}
}

// @Summary EditTask
// @Security ApiKeyAuth
// @Tags task
// @Description Изменить задачу по указанному id
// @ID edit-task
// @Accept  json
// @Produce  json
// @Param input body TaskRequest true "task info"
// @Param id   path      string  true  "Task ID (UUID)"
// @Success 200
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/{id} [patch]
func EditTask(taskProvider TaskProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Trace().Msg("get EditTask request")

		id := chi.URLParam(r, "id")

		uuid, err := uuid.Parse(id)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse path parameter")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid UUID"))
			return
		}

		var req TaskRequest
		err = render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse request")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err = taskProvider.Update(ctx, uuid, toModelTaskBody(req), req.CategoryIds)
		if err != nil {
			log.Err(err).Msg("Update, error from provider")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		render.Status(r, http.StatusOK)
	}
}

// @Summary GetTask
// @Security ApiKeyAuth
// @Tags task
// @Description Получить полное описание задачи по переданному id
// @ID get-task
// @Accept  json
// @Produce  json
// @Param id   path      string  true  "Task ID (UUID)"
// @Success 200 {object} TaskResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/{id} [get]
func GetTask(taskProvider TaskProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Trace().Msg("get GetTask request")

		id := chi.URLParam(r, "id")
		uuid, err := uuid.Parse(id)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse path parameter")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid UUID"))
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		task, err := taskProvider.GetByID(ctx, uuid)
		if err != nil {
			log.Err(err).Msg("GetByID, error from provider")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		render.JSON(w, r, toTaskResponse(task))
	}
}

// @Summary GetAllTasks
// @Security ApiKeyAuth
// @Tags task
// @Description Получить список коротких описаний задач
// @ID get-all-tasks
// @Accept  json
// @Produce  json
// @Param input body Pagination true "pagination info"
// @Success 200 {object} TasksList
// @Failure 400,401 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/all [post]
func GetAllTasks(taskProvider TaskProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Trace().Msg("get GetAllTasks request")

		var req Pagination
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse request")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		userId, ok := r.Context().Value(middleware.UserIDContextKey).(uuid.UUID)
		if !ok {
			log.Error().Msg("no uuid in context")
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, response.Error("unauthorized"))
			return
		}

		tasks, err := taskProvider.GetAll(ctx, userId, req.PageIndex, req.RecordsPerPage)
		if err != nil {
			log.Err(err).Msg("GetAll, error from provider")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		render.JSON(w, r, toTaskList(tasks))
	}
}

// @Summary ToggleReadinessTask
// @Security ApiKeyAuth
// @Tags task
// @Description Изменить статус готовности задачи. Если задача была готова, то станет неготовой или наоборот.
// @ID toggle-readiness-task
// @Accept  json
// @Produce  json
// @Param id   path      string  true  "Task ID (UUID)"
// @Success 200
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/{id}/readiness [post]
func ToggleReadinessTask(taskProvider TaskProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		uuid, err := uuid.Parse(id)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse path parameter")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid UUID"))
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err = taskProvider.ToggleDone(ctx, uuid)
		if err != nil {
			log.Err(err).Msg("ToggleDone, error from provider")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		render.Status(r, http.StatusOK)
	}
}

// @Summary DeleteTask
// @Security ApiKeyAuth
// @Tags task
// @Description Удалить задачу по переданному id
// @ID delete-task
// @Accept  json
// @Produce  json
// @Param id   path      string  true  "Task ID (UUID)"
// @Success 200
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/{id} [delete]
func DeleteTask(taskProvider TaskProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		uuid, err := uuid.Parse(id)
		if err != nil {
			log.Warn().Err(err).Msg("failed to parse path parameter")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("invalid UUID"))
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err = taskProvider.Delete(ctx, uuid)
		if err != nil {
			log.Err(err).Msg("Delete, error from provider")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		render.Status(r, http.StatusOK)
	}
}

func toModelTaskBody(task TaskRequest) *models.TaskBody {
	return &models.TaskBody{
		Title:       task.Title,
		Description: task.Description,
	}
}

func toTaskResponse(task *models.TaskFullInfo) *TaskResponse {
	categoryResponse := make([]CategoryResponse, len(task.Categories))
	for i, category := range task.Categories {
		categoryResponse[i] = CategoryResponse{
			ID: category.ID,
			CategoryBody: CategoryBody{
				Name: category.Name,
			},
		}
	}

	return &TaskResponse{
		TaskMeta: TaskMeta{
			ID:     task.ID,
			IsDone: task.IsDone,
		},
		TaskBody: TaskBody{
			Title:       task.Title,
			Description: task.Description,
		},
		CategoriesResponse: CategoriesResponse{
			Categories: categoryResponse,
		},
	}
}

func toTaskList(tasks []models.TaskShortInfo) TasksList {
	list := make([]TaskShortResponse, 0, len(tasks))
	for _, task := range tasks {
		list = append(list, TaskShortResponse{
			Title: task.Title,
			TaskMeta: TaskMeta{
				ID:     task.ID,
				IsDone: task.IsDone,
			},
		})
	}

	return TasksList{
		List: list,
	}
}
