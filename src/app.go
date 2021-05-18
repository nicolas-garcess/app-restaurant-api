package app

import "github.com/go-chi/chi/v5"

type App struct {
	Router chi.Router
}

func New() *App {
	return &App{
		Router: chi.NewRouter(),
	}
}
