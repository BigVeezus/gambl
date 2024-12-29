package controllers

import (
	"log"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword is used to encrypt the password before it is stored in the DB
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

// VerifyPassword checks the input password while verifying it with the passward in the DB.
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "login or passowrd is incorrect"
		check = false
	}

	return check, msg
}

func cleanUsername(username string) string {
	// Remove everything that is not an alphabet or a number
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	cleanedUsername := re.ReplaceAllString(username, "")

	// Optionally, convert to lowercase
	return strings.ToLower(cleanedUsername)
}

// Helper function to check if a user type is valid
func contains(validUserTypes []string, userType string) bool {
	for _, validType := range validUserTypes {
		if validType == userType {
			return true
		}
	}
	return false
}
