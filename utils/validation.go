package utils

import (
	"database/sql"
	"errors"
	"go-berry/models"
	"regexp"
	"strings"
)

func ValidateUserInput(user *models.User, db *sql.DB, isUpdate bool) error {
	if strings.TrimSpace(user.Name) == "" || strings.TrimSpace(user.Email) == "" || strings.TrimSpace(user.Password) == "" {
			return errors.New("name, email, and password are required")
	}

	if !isValidEmail(user.Email) {
			return errors.New("invalid email format")
	}

	if err := validatePasswordStrength(user.Password); err != nil {
			return err
	}

	if len(user.Name) < 3 || len(user.Name) > 50 {
			return errors.New("name must be between 3 and 50 characters")
	}

	if !isUpdate {
			emailExists, err := emailExists(user.Email, db)
			if err != nil {
					return err
			}
			if emailExists {
					return errors.New("email is already registered")
			}
	}

	return nil
}


func isValidEmail(email string) bool {
	const emailRegex = `(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func validatePasswordStrength(password string) error {
	var uppercase, lowercase, number, special bool
	if len(password) < 8 || len(password) > 100 {
		return errors.New("password must be between 8 and 100 characters")
	}
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			uppercase = true
		case 'a' <= char && char <= 'z':
			lowercase = true
		case '0' <= char && char <= '9':
			number = true
		case strings.ContainsRune("!@#$%^&*()-_=+[]{}|;:'\",.<>?/~`", char):
			special = true
		}
	}
	if !uppercase || !lowercase || !number || !special {
		return errors.New("password must include at least one uppercase letter, one lowercase letter, one number, and one special character")
	}
	return nil
}

func emailExists(email string, db *sql.DB) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
