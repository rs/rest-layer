package main

import (
	"log"
	"net/http"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
)

func main() {
	index := resource.NewIndex()

	// configure your resources

	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatalf("Invalid API configuration: %s", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fn := r.URL.Query().Get("callback")
		if fn != "" {
			w.Header().Set("Content-Type", "application/javascript")
			w.Write([]byte(";fn("))
		}
		api.ServeHTTP(w, r)
		if fn != "" {
			w.Write([]byte(");"))
		}
	})
	log.Fatal(http.ListenAndServe(":8080", handler))
}
