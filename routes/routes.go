package routes

import (
	"database/sql"
	"go-berry/handlers"

	"github.com/gorilla/mux"
)

func InitializeRoutes(r *mux.Router, db *sql.DB) {

	r.HandleFunc("/users", handlers.GetUsers(db)).Methods("GET")
	r.HandleFunc("/users/{id}", handlers.GetUser(db)).Methods("GET")
	r.HandleFunc("/users", handlers.CreateUser(db)).Methods("POST")
	r.HandleFunc("/users/{id}", handlers.UpdateUser(db)).Methods("PUT")
	r.HandleFunc("/users/{id}", handlers.DeleteUser(db)).Methods("DELETE")
}