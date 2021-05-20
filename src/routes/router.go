package routes

import (
	"fmt"
	"net/http"

	"api-sales/src/controller"

	"github.com/go-chi/chi/v5"
)

var SetUpServer = func(Port string) {
	router := chi.NewRouter()

	router.Get("/", controller.Index)
	router.Post("/upload-data", controller.UploadData)
	router.Get("/customers", controller.GetCustomers)
	router.Get("/customer/{id}", controller.GetCustomer)

	err := http.ListenAndServe(":"+Port, router)
	if err != nil {
		fmt.Println("No se pudo conectar al servidor")
	}
}
