package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
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

	handler := cors.Default().Handler(api)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
