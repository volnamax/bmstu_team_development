package handlers

import (
	"todolist/config"
	"todolist/internal/adapters"
	"todolist/internal/middleware"
	auth_utils "todolist/internal/pkg/authUtils"
	"todolist/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Handlers struct {
	db     *gorm.DB
	router *chi.Mux

	cfg *config.ServiceConfig
}

func NewHandlers(cfg *config.ServiceConfig, db *gorm.DB, router *chi.Mux) *Handlers {
	return &Handlers{
		cfg:    cfg,
		db:     db,
		router: router,
	}
}

func (h Handlers) InitHandlers() {
	h.initUserHandlers()
	tokenHandler := auth_utils.NewJWTTokenHandler()
	logger := logrus.New()
	authMiddleware := middleware.NewJwtAuthMiddleware(logger, h.cfg.JWTSecret, tokenHandler)
	h.router.Use(authMiddleware.MiddlewareFunc)
	h.initTaskHandlers()
	h.initCategoryHandlers()

}

func (h Handlers) initTaskHandlers() {
	taskRepo := repository.NewGormTaskRepository(h.db)
	taskUseCase := adapters.NewTaskAdapter(taskRepo)

	timeout := h.cfg.TaskTimeout

	userRepo := repository.NewUserRepositoryAdapter(h.db)
	jwtHandler := auth_utils.NewJWTTokenHandler()
	logger := logrus.New()
	userUseCase := adapters.NewAuthService(logger, userRepo, jwtHandler, h.cfg.JWTSecret)

	ownMiddleware := middleware.NewOwnershipMiddleware(*userUseCase, timeout)

	h.router.Route("/api/v1/task", func(r chi.Router) {

		r.With(ownMiddleware.CheckCategoriesMiddleware).Group(func(r chi.Router) {
			r.Post("/", CreateTask(taskUseCase, timeout))
		})

		r.With(ownMiddleware.CheckTaskMiddleware).Group(func(r chi.Router) {

			r.With(ownMiddleware.CheckCategoriesMiddleware).Group(func(r chi.Router) {
				r.Patch("/{id}", EditTask(taskUseCase, timeout))
			})
			r.Delete("/{id}", DeleteTask(taskUseCase, timeout))
			r.Post("/{id}/readiness", ToggleReadinessTask(taskUseCase, timeout))
			r.Get("/{id}", GetTask(taskUseCase, timeout))
		})

		r.Post("/all", GetAllTasks(taskUseCase, timeout))
	})
}

func (h Handlers) initUserHandlers() {

	timeout := h.cfg.TaskTimeout

	userRepo := repository.NewUserRepositoryAdapter(h.db)
	jwtHandler := auth_utils.NewJWTTokenHandler()
	logger := logrus.New()
	userUseCase := adapters.NewAuthService(logger, userRepo, jwtHandler, h.cfg.JWTSecret)

	h.router.Route("/api/v1", func(r chi.Router) {
		r.Post("/sign-in", SignIn(userUseCase, timeout))
		r.Post("/sign-up", SignUp(userUseCase, timeout))
		r.Delete("/user", DeleteUser(userUseCase, timeout))
	})
}

func (h Handlers) initCategoryHandlers() {

	timeout := h.cfg.TaskTimeout

	categoryRepo := repository.NewCategoryRepositoryAdapter(h.db)
	categoryUseCase := adapters.NewCategoryAdapter(categoryRepo)
	h.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/category", GetCategories(categoryUseCase, timeout))
		r.Post("/category", CreateCategory(categoryUseCase, timeout))
		r.Delete("/category", CreateCategory(categoryUseCase, timeout))
	})
}
