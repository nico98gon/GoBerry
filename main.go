package main

import (
	"log"
	"net/http"

	"go-berry/config"
	"go-berry/middleware"
	"go-berry/routes"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main () {
	// connnect to database
	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// initialize routes
	r := mux.NewRouter()
	routes.InitializeRoutes(r, db)

	// start server
	log.Fatal(http.ListenAndServe(":8080", middleware.JsonContentMiddleware(r)))
}

