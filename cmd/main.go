package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	zlog "github.com/rs/zerolog/log"

	"github.com/avast/retry-go"
	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"todolist/config"
	_ "todolist/docs"
	"todolist/internal/api/handlers"

	"github.com/rs/zerolog"
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
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.BuildConfig()
	if err != nil {
		log.Panic("failed parse config: ", err)
	}

	db, err := connectWithRetry(cfg.PostgresConfig.String())
	if err != nil {
		log.Panic("Could not connect to DB after retries: ", err)
	}
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	r := chi.NewRouter()
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	handlersBuilder := handlers.NewHandlers(&cfg.ServiceConfig, db, r)
	handlersBuilder.InitHandlers()

	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: r,
	}

	go func() {
		zlog.Trace().Msg("starting server")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panic("failed to start server")
		}
	}()

	<-done
	zlog.Trace().Msg("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zlog.Error().Msg("failed to stop server")
		return
	}

	zlog.Trace().Msg("server stopped")
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
