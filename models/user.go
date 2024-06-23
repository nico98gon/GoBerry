package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Email       string                 `json:"email"`
	Username    string                 `json:"username,omitempty"`
	Password    string                 `json:"password"`
	// Phone       string                 `json:"phone,omitempty"`
	// Address     string                 `json:"address,omitempty"`
	// DateOfBirth time.Time              `json:"date_of_birth,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastLogin   time.Time              `json:"last_login,omitempty"`
	IsActive    bool                   `json:"is_active"`
	Groups      []Group                `json:"groups,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type Group struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type PaginatedResponse struct {
	Users      []User `json:"users"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalUsers int    `json:"total_users"`
}
