package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("salut !"))
		if err != nil {
			log.Println(err)
			return
		}
	})

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Println(err)
		return
	}
}

func Add(a, b int) int {
	return a + b
}
