package config

import (
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/joho/godotenv"
)

func CLoudinaryInstance() *cloudinary.Cloudinary {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file, proceeding with environment variables")
	}

	CLD_NAME := os.Getenv("CLD_NAME")
	CLD_KEY := os.Getenv("CLD_API_KEY")
	CLD_SECRET := os.Getenv("CLD_SECRET")

	if CLD_NAME == "" || CLD_KEY == "" || CLD_SECRET == "" {
		// Use a default value or handle the case where MONGO_URL is not set
		log.Fatal("Cloudinary environment variable is not set")
	}

	// Add your Cloudinary product environment credentials.
	cld, err := cloudinary.NewFromParams(CLD_NAME, CLD_KEY, CLD_SECRET)

	if err != nil {
		log.Fatal("Error creating cloudinary")
	}

	return cld
}
func CloudinaryFolder() string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	CLD_FOLDER := os.Getenv("CLD_SECRET")

	return CLD_FOLDER
}

var CloudinaryClient *cloudinary.Cloudinary = CLoudinaryInstance()
