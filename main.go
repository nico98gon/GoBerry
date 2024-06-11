package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User	struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

func main () {
	// connnect to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")

	if err != nil {
		log.Fatal(err)
	}

	// create router
	r := mux.NewRouter()
	r.HandleFunc("/users", getUsers(db)).Methods("GET")
	r.HandleFunc("/users/{id}", getUser(db)).Methods("GET")
	r.HandleFunc("/users", createUser(db)).Methods("POST")
	r.HandleFunc("/users/{id}", updateUser(db)).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUser(db)).Methods("DELETE")

	// start server
	log.Fatal(http.ListenAndServe(":8080", jsonContentMiddleware(r)))
}

func jsonContentMiddleware (next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// get all users
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		users := []User{}
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
				// TODO: fix error handling
				log.Fatal(err)
			}
			users = append(users, user)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

// get user by id
func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var user User
		err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// create user
func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			// TODO: fix error handling
			log.Fatal(err)
		}
		_, err := db.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", user.Name, user.Email)
		if err != nil {
			// TODO: fix error handling
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusCreated)
	}
}

// update user
func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		json.NewDecoder(r.Body).Decode(&user)

		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", user.Name, user.Email, id)
		if err != nil {
			// TODO: fix error handling
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}
}

// delete user
func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		name := vars["name"]
		email := vars["email"]

		_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			// TODO: fix error handling
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode("User " + name + " deleted id: " + id + "email: " + email)
	}
}
