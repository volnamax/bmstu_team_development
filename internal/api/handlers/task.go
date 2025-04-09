package handlers

import (
	"net/http"

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
func CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
