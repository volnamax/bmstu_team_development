package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	_ "todolist/docs"

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
	r := chi.NewRouter()
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	http.ListenAndServe(":8080", r)
}