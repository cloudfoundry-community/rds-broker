package main

import (
	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"

	"fmt"
	"os"
)

type RDS struct {
	Url      string
	Username string
	Password string
	DbName   string
	Sslmode  string
	Port     string
}

func LoadRDS() *RDS {
	rds := RDS{}
	rds.Url = os.Getenv("DB_URL")
	rds.Username = os.Getenv("DB_USER")
	rds.Password = os.Getenv("DB_PASS")
	rds.DbName = os.Getenv("DB_NAME")
	rds.Sslmode = "disable"

	if os.Getenv("DB_PORT") != "" {
		rds.Port = os.Getenv("DB_PORT")
	} else {
		rds.Port = "5432"
	}

	return &rds
}

func main() {
	rds := LoadRDS()

	m := App(rds, "prod")

	m.Run()
}

func App(rds *RDS, env string) *martini.ClassicMartini {
	err := DBInit(rds, env)
	if err != nil {
		fmt.Println("There was an error with the DB")
		return nil
	}

	m := martini.Classic()

	username := os.Getenv("AUTH_USER")
	password := os.Getenv("AUTH_PASS")

	m.Use(auth.Basic(username, password))
	m.Use(render.Renderer())

	m.Map(&DB)
	m.Map(rds)

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
