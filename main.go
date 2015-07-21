package main

import (
	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"

	"log"
	"os"
)

func main() {
	var settings Settings

	// Load settings from environment
	if err := settings.LoadFromEnv(); err != nil {
		log.Println("There was an error loading settings")
		log.Println(err)
		return
	}

	DB, err := InternalDBInit(settings.DbConfig)
	if err != nil {
		log.Println("There was an error with the DB. Error: " + err.Error())
		return
	}

	// Try to connect and create the app.
	if m := App(&settings, DB); m != nil {
		log.Println("Starting app...")
		m.Run()
	} else {
		log.Println("Unable to setup application. Exiting...")
	}
}

func App(settings *Settings, DB *gorm.DB) *martini.ClassicMartini {

	m := martini.Classic()

	username := os.Getenv("AUTH_USER")
	password := os.Getenv("AUTH_PASS")

	m.Use(auth.Basic(username, password))
	m.Use(render.Renderer())

	m.Map(DB)
	m.Map(settings)

	log.Println("Loading Routes")

	// Serve the catalog with services and plans
	m.Get("/v2/catalog", func(r render.Render) {
		services := BuildCatalog()
		catalog := map[string]interface{}{
			"services": services,
		}
		r.JSON(200, catalog)
	})

	// Create the service instance (cf create-service-instance)
	m.Put("/v2/service_instances/:id", CreateInstance)

	// Bind the service to app (cf bind-service)
	m.Put("/v2/service_instances/:instance_id/service_bindings/:id", BindInstance)

	// Unbind the service from app
	m.Delete("/v2/service_instances/:instance_id/service_bindings/:id", func(p martini.Params, r render.Render) {
		var emptyJson struct{}
		r.JSON(200, emptyJson)
	})

	// Delete service instance
	m.Delete("/v2/service_instances/:id", DeleteInstance)

	return m
}
