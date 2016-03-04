package main

import (
	"log"
	"net/http"

	"github.com/rs/rest-layer/examples/auth_service/models"

	"github.com/rs/rest-layer-mem"
	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
)


func main() {
	index := resource.NewIndex()

	users := index.Bind("users", resource.New(models.UserSchema, mem.NewHandler(), resource.Conf{
		AllowedModes:           resource.ReadWrite,
		PaginationDefaultLimit: 50,

	}))

	models.SetAuthUserResource(users)

	index.Bind("auth", resource.New(models.AuthSchema, mem.NewHandler(), resource.Conf{
		AllowedModes: []resource.Mode{resource.Create},
	}))

	api, err := rest.NewHandler(index)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", api)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
