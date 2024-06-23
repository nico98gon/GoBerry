package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-berry/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetAllUsers(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error to create the mock database: %v", err)
	}
	defer db.Close()

	// Generate user data
	userCount := 10
	expectedUsers := make([]models.User, userCount)
	rows := sqlmock.NewRows([]string{"id", "name", "email"})
	for i := 1; i <= userCount; i++ {
		user := models.User{
			ID:    uuid.New(),
			Name:  faker.Name(),
			Email: faker.Email(),
		}
		expectedUsers[i-1] = user
		rows.AddRow(user.ID, user.Name, user.Email)
	}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(userCount))

	mock.ExpectQuery("SELECT id, name, email FROM users LIMIT (.+) OFFSET (.+)").
		WithArgs(10, 0).
		WillReturnRows(rows)

	// Create a simulated HTTP request
	req, err := http.NewRequest("GET", "/users?page=1&limit=10", nil)
	if err != nil {
		t.Fatalf("Error to create the HTTP request: %v", err)
	}

	rr := httptest.NewRecorder()

	handler := GetAllUsers(db)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Should return status 200 OK")

	expectedResponse := models.PaginatedResponse{
		Users:      expectedUsers,
		Page:       1,
		Limit:      10,
		TotalUsers: userCount,
	}

	var actualResponse models.PaginatedResponse
	err = json.NewDecoder(rr.Body).Decode(&actualResponse)
	if err != nil {
		t.Fatalf("Error encoding the response: %v", err)
	}

	// Compare only the lengths of the slices to ensure the response structure is correct
	assert.Equal(t, len(expectedResponse.Users), len(actualResponse.Users), "The number of users should match")
	assert.Equal(t, expectedResponse.Page, actualResponse.Page, "The page should match")
	assert.Equal(t, expectedResponse.Limit, actualResponse.Limit, "The limit should match")
	assert.Equal(t, expectedResponse.TotalUsers, actualResponse.TotalUsers, "The total number of users should match")

	// Compare each user in the list
	for i := range expectedResponse.Users {
		assert.Equal(t, expectedResponse.Users[i].ID, actualResponse.Users[i].ID, "User IDs should match")
		assert.Equal(t, expectedResponse.Users[i].Name, actualResponse.Users[i].Name, "User names should match")
		assert.Equal(t, expectedResponse.Users[i].Email, actualResponse.Users[i].Email, "User emails should match")
	}
}


func TestGetUser(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating the mock database: %v", err)
	}
	defer db.Close()

	// Expected user data
	userID := uuid.New()
	expectedUser := models.User{
		ID:        userID,
		Name:      faker.Name(),
		Email:     faker.Email(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	mock.ExpectQuery("SELECT id, name, email, created_at, updated_at, is_active FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at", "is_active"}).
			AddRow(expectedUser.ID, expectedUser.Name, expectedUser.Email, expectedUser.CreatedAt, expectedUser.UpdatedAt, expectedUser.IsActive))

	// Create a simulated HTTP request
	req, err := http.NewRequest("GET", "/users/"+userID.String(), nil)
	if err != nil {
		t.Fatalf("Error creating the HTTP request: %v", err)
	}

	req = mux.SetURLVars(req, map[string]string{"id": userID.String()})

	rr := httptest.NewRecorder()

	handler := GetUser(db)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Should return status 200 OK")

	var actualUser models.User
	err = json.NewDecoder(rr.Body).Decode(&actualUser)
	if err != nil {
		t.Fatalf("Error decoding the response: %v", err)
	}

	assert.Equal(t, expectedUser.ID, actualUser.ID, "User IDs should match")
	assert.Equal(t, expectedUser.Name, actualUser.Name, "User names should match")
	assert.Equal(t, expectedUser.Email, actualUser.Email, "User emails should match")
	assert.WithinDuration(t, expectedUser.CreatedAt, actualUser.CreatedAt, time.Second, "User created_at timestamps should match")
	assert.WithinDuration(t, expectedUser.UpdatedAt, actualUser.UpdatedAt, time.Second, "User updated_at timestamps should match")
	assert.Equal(t, expectedUser.IsActive, actualUser.IsActive, "User is_active status should match")

	// Ensure the password field is empty in the response
	assert.Equal(t, "", actualUser.Password, "Password field should be empty")
}



func TestCreateUser(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating the mock database: %v", err)
	}
	defer db.Close()

	userInput := models.User{
		Name:     faker.Name(),
		Email:    faker.Email(),
		Password: "StrongP@ssw0rd",
	}

	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO users").
		WithArgs(sqlmock.AnyArg(), userInput.Name, userInput.Email, sqlmock.AnyArg(), now, now, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	body, err := json.Marshal(userInput)
	if err != nil {
		t.Fatalf("Error marshaling user input: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	rr := httptest.NewRecorder()

	handler := CreateUser(db)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rr.Code)
		t.Logf("Response body: %s", rr.Body.String())
		return
	}

	// Check the response body
	var responseUser models.User
	if err = json.NewDecoder(rr.Body).Decode(&responseUser); err != nil {
		t.Fatalf("Error decoding response body: %v", err)
	}

	// Ensure the password is not returned in the response
	assert.Empty(t, responseUser.Password)
	assert.Equal(t, userInput.Name, responseUser.Name)
	assert.Equal(t, userInput.Email, responseUser.Email)
	// Since the ID is generated in the handler, we don't compare with newUUID here
	assert.True(t, responseUser.IsActive)
	assert.WithinDuration(t, now, responseUser.CreatedAt, time.Second)
	assert.WithinDuration(t, now, responseUser.UpdatedAt, time.Second)

	// Ensure all expectations were met
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %v", err)
	}
}


func TestUpdateUser(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating the mock database: %v", err)
	}
	defer db.Close()

	userID := uuid.New()
	expectedUser := models.User{
		ID:        userID,
		Name:      faker.Name(),
		Email:     faker.Email(),
		UpdatedAt: time.Now(),
	}

	// Mock validation function
	// utils.ValidateUserInput = func(user *models.User, db *sql.DB, isUpdate bool) error {
	// 	return nil
	// }

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET name = \\$1, email = \\$2, updated_at = \\$3 WHERE id = \\$4").
		WithArgs(expectedUser.Name, expectedUser.Email, sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Create a simulated HTTP request
	body, err := json.Marshal(expectedUser)
	if err != nil {
		t.Fatalf("Error marshaling user: %v", err)
	}

	req, err := http.NewRequest("PUT", "/users/"+userID.String(), bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Error creating the HTTP request: %v", err)
	}

	req = mux.SetURLVars(req, map[string]string{"id": userID.String()})

	rr := httptest.NewRecorder()

	handler := UpdateUser(db)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Should return status 200 OK")

	var actualUser models.User
	err = json.NewDecoder(rr.Body).Decode(&actualUser)
	if err != nil {
		t.Fatalf("Error decoding the response: %v", err)
	}

	assert.Equal(t, expectedUser.ID, actualUser.ID, "User IDs should match")
	assert.Equal(t, expectedUser.Name, actualUser.Name, "User names should match")
	assert.Equal(t, expectedUser.Email, actualUser.Email, "User emails should match")
	assert.WithinDuration(t, expectedUser.UpdatedAt, actualUser.UpdatedAt, time.Second, "User updated_at timestamps should match")

	// Ensure the password field is empty in the response
	assert.Equal(t, "", actualUser.Password, "Password field should be empty")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}


func TestDeleteUser(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating the mock database: %v", err)
	}
	defer db.Close()

	userID := uuid.New()
	expectedUser := models.User{
		ID:    userID,
		Name:  faker.Name(),
		Email: faker.Email(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT name, email FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"name", "email"}).AddRow(expectedUser.Name, expectedUser.Email))

	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Create a simulated HTTP request
	req, err := http.NewRequest("DELETE", "/users/"+userID.String(), nil)
	if err != nil {
		t.Fatalf("Error creating the HTTP request: %v", err)
	}

	req = mux.SetURLVars(req, map[string]string{"id": userID.String()})

	rr := httptest.NewRecorder()

	handler := DeleteUser(db)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Should return status 200 OK")

	var actualResponse map[string]string
	err = json.NewDecoder(rr.Body).Decode(&actualResponse)
	if err != nil {
		t.Fatalf("Error decoding the response: %v", err)
	}

	expectedMessage := fmt.Sprintf("User %s with ID %s and email %s deleted successfully", expectedUser.Name, userID, expectedUser.Email)
	assert.Equal(t, expectedMessage, actualResponse["message"], "Response message should match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %v", err)
	}
}
