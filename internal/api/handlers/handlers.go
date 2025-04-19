package handlers

import (
	"net/http"
	"todolist/config"
	"todolist/internal/adapters"
	"todolist/internal/middleware"
	auth_utils "todolist/internal/pkg/authUtils"
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
		cfg:    cfg,
		db:     db,
		router: router,
	}
}

func (h Handlers) InitHandlers() {
	h.initUserHandlers()
	h.initTaskHandlers()
	h.initCategoryHandlers()

}

func (h Handlers) initTaskHandlers() {
	taskRepo := repository.NewGormTaskRepository(h.db)
	taskUseCase := adapters.NewTaskAdapter(taskRepo)

	timeout := h.cfg.TaskTimeout

	userRepo := repository.NewUserRepositoryAdapter(h.db)
	jwtHandler := auth_utils.NewJWTTokenHandler()
	userUseCase := adapters.NewAuthService(userRepo, jwtHandler, h.cfg.JWTSecret)

	ownMiddleware := middleware.NewOwnershipMiddleware(*userUseCase, timeout)

	tokenHandler := auth_utils.NewJWTTokenHandler()
	authMiddleware := middleware.NewJwtAuthMiddleware(h.cfg.JWTSecret, tokenHandler)

	h.router.Route("/api/v1/task", func(r chi.Router) {
		r.With(authMiddleware.MiddlewareFunc).Group(func(r chi.Router) {

			r.With(ownMiddleware.CheckCategoriesMiddleware).Group(func(r chi.Router) {
				r.Post("/", CreateTask(taskUseCase, timeout))
			})

			r.Route("/{id}", func(r chi.Router) {
				r.Use(
					func(next http.Handler) http.Handler {
						return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							// Force Chi to bind {id} by calling a no-op handler
							chi.URLParam(r, "id")
							next.ServeHTTP(w, r)
						})
					},
					ownMiddleware.CheckTaskMiddleware, // Now {id} is available here
				)

				r.Group(func(r chi.Router) {
					r.With(ownMiddleware.CheckCategoriesMiddleware).Patch("/", EditTask(taskUseCase, timeout))
				})

				r.Delete("/", DeleteTask(taskUseCase, timeout))
				r.Post("/readiness", ToggleReadinessTask(taskUseCase, timeout))
				r.Get("/", GetTask(taskUseCase, timeout))
			})

			r.Post("/all", GetAllTasks(taskUseCase, timeout))
		})
	})
}

func (h Handlers) initUserHandlers() {

	timeout := h.cfg.TaskTimeout

	userRepo := repository.NewUserRepositoryAdapter(h.db)
	jwtHandler := auth_utils.NewJWTTokenHandler()
	userUseCase := adapters.NewAuthService(userRepo, jwtHandler, h.cfg.JWTSecret)

	tokenHandler := auth_utils.NewJWTTokenHandler()
	authMiddleware := middleware.NewJwtAuthMiddleware(h.cfg.JWTSecret, tokenHandler)

	h.router.Route("/api/v1", func(r chi.Router) {
		r.Post("/sign-in", SignIn(userUseCase, timeout))
		r.Post("/sign-up", SignUp(userUseCase, timeout))
		r.With(authMiddleware.MiddlewareFunc).Group(func(r chi.Router) {
			r.Delete("/user", DeleteUser(userUseCase, timeout))
		})
	})
}

func (h Handlers) initCategoryHandlers() {

	timeout := h.cfg.TaskTimeout

	categoryRepo := repository.NewCategoryRepositoryAdapter(h.db)
	categoryUseCase := adapters.NewCategoryAdapter(categoryRepo)
	tokenHandler := auth_utils.NewJWTTokenHandler()

	authMiddleware := middleware.NewJwtAuthMiddleware(h.cfg.JWTSecret, tokenHandler)
	h.router.Route("/api/v1/category", func(r chi.Router) {
		r.With(authMiddleware.MiddlewareFunc).Group(func(r chi.Router) {
			r.Post("/all", GetCategories(categoryUseCase, timeout))
			r.Post("/", CreateCategory(categoryUseCase, timeout))
			r.Delete("/{id}", DeleteCategory(categoryUseCase, timeout))
		})
	})
}
