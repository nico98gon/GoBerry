package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"go-berry/models"

	"github.com/gorilla/mux"
)

// get all users
func GetUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM users")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		users := []models.User{}
		for rows.Next() {
			var user models.User
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
func GetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var user models.User
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
func CreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
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
func UpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
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
func DeleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var user models.User
		err := db.QueryRow("SELECT name, email FROM users WHERE id = $1", id).Scan(&user.Name, &user.Email)
		if err != nil {
			// TODO: handle error
			log.Fatal(err)
		}

		_, err = db.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			// TODO: handle error
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode("User " + user.Name + " deleted id: " + id + " email: " + user.Email)
	}
}
