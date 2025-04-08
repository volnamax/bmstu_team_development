package handlers

import "net/http"

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"` 
}

// @Summary CreateTask
// @Security ApiKeyAuth
// @Tags task
// @Description create task
// @ID create-task
// @Accept  json
// @Produce  json
// @Param input body CreateTaskRequest true "task info"
// @Success 200
// @Failure 400,404 {object} response.Response
// @Failure 500 {object} response.Response
// @Failure default {object} response.Response
// @Router /api/v1/task [post]
func CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}