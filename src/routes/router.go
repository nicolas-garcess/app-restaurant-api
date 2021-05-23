package routes

import (
	"fmt"
	"net/http"

	"api-sales/src/controller"

	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
)

var SetUpServer = func(Port string) {
	router := chi.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	router.Use(c.Handler)

	router.Get("/", controller.Index)
	router.Post("/data", controller.UploadData)
	router.Get("/customers", controller.GetCustomers)
	router.Get("/customer/{id}", controller.GetCustomer)

	err := http.ListenAndServe(":"+Port, router)
	if err != nil {
		fmt.Println(err)
	}
}
