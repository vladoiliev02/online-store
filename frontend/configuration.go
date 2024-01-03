package frontend

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Init(router chi.Router) {
	fs := http.FileServer(http.Dir("./static"))

	router.Get("/store/products/{productId}", serveFile("static/product.html"))
	router.Get("/store/users/{userId}", serveFile("static/user.html"))
	router.Get("/store/orders/{orderId}", serveFile("static/order.html"))
	router.Mount("/store", http.StripPrefix("/store/", fs))
}

func serveFile(fileName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, fileName)
	}
}
