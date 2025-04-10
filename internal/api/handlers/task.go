package handlers

import (
	"context"
	"net/http"
	"time"
	"todolist/internal/models"
	"todolist/internal/pkg/response"

	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type TaskBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TaskMeta struct {
	Id     uuid.UUID `json:"id"`
	IsDone bool      `json:"is_done"`
}

type TaskRequest struct {
	TaskBody
	CategoryIds []CategoryId `json:"category_ids"`
}

type TaskResponse struct {
	TaskMeta
	TaskBody
	CategoriesResponse
}

type TasksList struct {
	List []struct {
		TaskMeta
		Title string `json:"title"`
	} `json:"list"`
}

type TaskProvider interface {
	CreateTask(ctx context.Context, body *models.TaskBody) error
	Update(ctx context.Context, id uuid.UUID, body *models.TaskBody) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.TaskBody, error)
	GetAll(ctx context.Context, pageIndex, recordsPerPage int) ([]models.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ToggleDone(ctx context.Context, id uuid.UUID) error
}

// @Summary CreateTask
// @Security ApiKeyAuth
// @Tags task
// @Description create task
// @ID create-task
// @Accept  json
// @Produce  json
// @Param input body TaskRequest true "task info"
// @Success 200
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task [post]
func CreateTask(taskProvider TaskProvider, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TaskRequest
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error(err.Error()))
			return
		}

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err = taskProvider.CreateTask(ctx, toModelTaskBody(req))
		if err != nil {
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
// @Description edit task
// @ID edit-task
// @Accept  json
// @Produce  json
// @Param input body TaskRequest true "task info"
// @Param id   path      string  true  "Task ID (UUID)"
// @Success 200
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/{id} [put]
func EditTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// @Summary GetTask
// @Security ApiKeyAuth
// @Tags task
// @Description get task
// @ID get-task
// @Accept  json
// @Produce  json
// @Param id   path      string  true  "Task ID (UUID)"
// @Success 200 {object} TaskResponse
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/{id} [get]
func GetTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// @Summary GetAllTasks
// @Security ApiKeyAuth
// @Tags task
// @Description get all tasks
// @ID get-all-tasks
// @Accept  json
// @Produce  json
// @Param input body Pagination true "pagination info"
// @Success 200 {object} TasksList
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/all [post]
func GetAllTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// @Summary ToggleReadinessTask
// @Security ApiKeyAuth
// @Tags task
// @Description toggle readiness task
// @ID toggle-readiness-task
// @Accept  json
// @Produce  json
// @Param id   path      string  true  "Task ID (UUID)"
// @Success 200
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/{id} [post]
func ToggleReadinessTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// @Summary DeleteTask
// @Security ApiKeyAuth
// @Tags task
// @Description delete task
// @ID delete-task
// @Accept  json
// @Produce  json
// @Param id   path      string  true  "Task ID (UUID)"
// @Success 200
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task/{id} [delete]
func DeleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func toModelTaskBody(task TaskRequest) *models.TaskBody {
	return &models.TaskBody{
		Title:       task.Title,
		Description: task.Description,
	}
}
