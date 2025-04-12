package handlers

import (
	"todolist/config"
	"todolist/internal/adapters"
	"todolist/internal/repository"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Handlers struct {
	db     *gorm.DB
	router *chi.Mux

	cfg *config.ServiceConfig
}

func NewHandlers(cfg *config.ServiceConfig, db *gorm.DB, router *chi.Mux) *Handlers {
	return &Handlers{
		db:     db,
		router: router,
	}
}

func (h Handlers) InitHandlers() {
	h.initTaskHandlers()
}

func (h Handlers) initTaskHandlers() {
	taskRepo := repository.NewGormTaskRepository(h.db)
	taskUseCase := adapters.NewTaskAdapter(taskRepo)

	timeout := h.cfg.TaskTimeout

	h.router.Post("/api/v1/task", CreateTask(taskUseCase, timeout))
	h.router.Patch("/api/v1/task/{id}", EditTask(taskUseCase, timeout))
	h.router.Get("/api/v1/task/{id}", GetTask(taskUseCase, timeout))
	h.router.Post("/api/v1/task/all", GetAllTasks(taskUseCase, timeout))
	h.router.Post("/api/v1/task/{id}/readiness", ToggleReadinessTask(taskUseCase, timeout))
	h.router.Delete("/api/v1/task/{id}", DeleteTask(taskUseCase, timeout))
}
