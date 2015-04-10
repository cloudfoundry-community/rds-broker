package main

import (
	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"

	"os"
)

func main() {
	m := App()

	m.Run()
}

func App() *martini.ClassicMartini {
	m := martini.Classic()

	username := os.Getenv("AUTH_USER")
	password := os.Getenv("AUTH_PASS")

	m.Use(auth.Basic(username, password))
	m.Use(render.Renderer())

	// Serve the catalog with services and plans
	m.Get("/v2/catalog", func(r render.Render) {
		services := BuildCatalog()
		catalog := map[string]interface{}{
			"services": services,
		}
		r.JSON(200, catalog)
	})

	// Create the service instance (cf create-service-instance)
	m.Put("/v2/service_instances/:id", func(p martini.Params, r render.Render) {
		var emptyJson struct{}
		r.JSON(201, emptyJson)
	})

	// Bind the service to app (cf bind-service)
	m.Put("/v2/service_instances/:instance_id/service_bindings/:id", func(p martini.Params, r render.Render) {
		credentials := map[string]string{
			"uri": "postgres://some@thing:my.database.com/" + p["id"],
		}
		response := map[string]interface{}{
			"credentials": credentials,
		}
		r.JSON(201, response)
	})

	// Unbind the service from app
	m.Delete("/v2/service_instances/:instance_id/service_bindings/:id", func(p martini.Params, r render.Render) {
		var emptyJson struct{}
		r.JSON(200, emptyJson)
	})

	// Delete service instance
	m.Delete("/v2/service_instances/:id", func(p martini.Params, r render.Render) {
		var emptyJson struct{}
		r.JSON(200, emptyJson)
	})

	return m
}
