package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-berry/models"
	"go-berry/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// handles GET requests to retrieve all users with pagination
func GetAllUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get pagination parameters from query
		pageParam := r.URL.Query().Get("page")
		limitParam := r.URL.Query().Get("limit")

		// Set default values if parameters are not provided
		page := 1
		limit := 10

		if pageParam != "" {
			p, err := strconv.Atoi(pageParam)
			if err == nil && p > 0 {
				page = p
			}
		}

		if limitParam != "" {
			l, err := strconv.Atoi(limitParam)
			if err == nil && l > 0 {
				limit = l
			}
		}

		offset := (page - 1) * limit

		// Fetch total number of users for pagination metadata
		var totalUsers int
		err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
		if err != nil {
			log.Printf("Error counting users: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Fetch users from database
		rows, err := db.Query("SELECT id, name, email FROM users LIMIT $1 OFFSET $2", limit, offset)
		if err != nil {
			log.Printf("Error querying users: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		users := []models.User{}
		for rows.Next() {
			var user models.User
			if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
				log.Printf("Error scanning user: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}

		if err := rows.Err(); err != nil {
			log.Printf("Error iterating over rows: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Respond with JSON including pagination metadata
		response := models.PaginatedResponse{
			Users:      users,
			Page:       page,
			Limit:      limit,
			TotalUsers: totalUsers,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// handles GET requests to retrieve a single user by ID
func GetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var user models.User
		err := db.QueryRow("SELECT id, name, email, created_at, updated_at, is_active FROM users WHERE id = $1", id).Scan(
			&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.IsActive,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				log.Printf("Error querying user: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Do not include the password in the response
		user.Password = ""

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(user); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// handles POST requests to create a new user
func CreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if err := utils.ValidateUserInput(&user, db, false); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		match := utils.CheckPasswordHash(user.Password, hashedPassword)
		if !match {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		user.Password = hashedPassword

		user.ID = uuid.New()

		now := time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now
		user.IsActive = true

		// Use a transaction for atomicity
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Error beginning transaction: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(
			"INSERT INTO users (id, name, email, password, created_at, updated_at, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			user.ID, user.Name, user.Email, user.Password, user.CreatedAt, user.UpdatedAt, user.IsActive,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("Error inserting user: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("Error committing transaction: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Do not include the password in the response
		user.Password = ""

		// Respond with JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

// handles PUT requests to update an existing user
func UpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if err := utils.ValidateUserInput(&user, db, true); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
}

		vars := mux.Vars(r)
		id := vars["id"]

		now := time.Now()
		user.UpdatedAt = now

		// Use a transaction for atomicity
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Error beginning transaction: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(
			"UPDATE users SET name = $1, email = $2, updated_at = $3 WHERE id = $4",
			user.Name, user.Email, user.UpdatedAt, id,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("Error updating user: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("Error committing transaction: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Do not include the password in the response
		user.Password = ""

		// Respond with JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}


// handles DELETE requests to delete an existing user
func DeleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var user models.User

		// Use a transaction for atomicity
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Error beginning transaction: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Fetch the user details before deletion
		err = tx.QueryRow("SELECT name, email FROM users WHERE id = $1", id).Scan(&user.Name, &user.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				log.Printf("Error querying user: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			tx.Rollback()
			return
		}

		_, err = tx.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			log.Printf("Error deleting user: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			log.Printf("Error committing transaction: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"message": fmt.Sprintf("User %s with ID %s and email %s deleted successfully", user.Name, id, user.Email),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

