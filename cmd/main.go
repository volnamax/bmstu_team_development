package main

import (
	"log"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"todolist/config"
	_ "todolist/docs"
	"todolist/internal/api/handlers"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Plan&Do API
// @version 1.0
// @description API Server for todo list

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	cfg, err := config.BuildConfig()
	if err != nil {
		log.Panic("failed parse config: %v", err)
	}

	db, err := connectWithRetry(cfg.PostgresConfig.String())
	if err != nil {
		log.Panic("Could not connect to DB after retries: %v", err)
	}

	r := chi.NewRouter()
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	handlersBuilder := handlers.NewHandlers(&cfg.ServiceConfig, db, r)
	handlersBuilder.InitHandlers()

	http.ListenAndServe(":8080", r)
}

func connectWithRetry(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	err := retry.Do(
		func() error {
			var err error
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
			if err != nil {
				log.Printf("DB connection failed: %v. Retrying...", err)
			}
			return err
		},
		retry.Attempts(5),
		retry.Delay(1*time.Second),
		retry.MaxDelay(10*time.Second),
		retry.DelayType(retry.BackOffDelay),
	)
	return db, err
}
