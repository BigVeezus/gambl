package controllers

import (
	"fmt"
	"gambl/core/user"
	helper "gambl/helpers"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/go-playground/validator/v10"

	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService user.UserService
	logger      *log.Logger
}

func NewUserController(us user.UserService, l *log.Logger) *UserController {
	return &UserController{
		userService: us,
		logger:      l,
	}
}

func (uc *UserController) CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Init: create user controller")

		var user user.User

		// Bind incoming JSON payload to struct, no extra allocations or memory copy
		if err := c.BindJSON(&user); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				var errorMessages []string
				for _, e := range validationErrors {
					errorMessages = append(errorMessages, fmt.Sprintf(
						"Field: %s, Error: %s, Value: %v",
						e.Field(),
						e.Tag(),
						e.Value(),
					))
				}

				// Log errors only if it's a significant issue
				uc.logger.Printf("Validation errors:\n%s", strings.Join(errorMessages, "\n"))

				// Return error response in a single pass without extra allocations
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation failed",
					"details": errorMessages,
				})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Ensure user creation failure doesn't cause memory leaks or excessive allocations
		if err := uc.userService.CreateUser(c.Request.Context(), &user); err != nil {
			uc.logger.Printf("Failed to create user, error: %v", err)
			// Minimize memory overhead in the error response
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Return a minimal response to ensure we're not unnecessarily allocating large objects
		c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
	}
}

func (uc *UserController) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginRequest struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		// Bind incoming JSON payload to struct, ensuring no unnecessary allocations
		if err := c.BindJSON(&loginRequest); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				var errorMessages []string
				for _, e := range validationErrors {
					errorMessages = append(errorMessages, fmt.Sprintf(
						"Field: %s, Error: %s, Value: %v",
						e.Field(),
						e.Tag(),
						e.Value(),
					))
				}

				// Log only significant issues (e.g., validation errors)
				uc.logger.Printf("Validation errors:\n%s", strings.Join(errorMessages, "\n"))

				// Return error response in a single pass without extra allocations
				c.JSON(http.StatusBadGateway, gin.H{
					"error":   "Validation failed",
					"details": errorMessages,
				})
				return
			}

			// Handle generic invalid request errors
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Fetch user by email (assuming a GetUserByEmail method exists)
		user, err := uc.userService.GetUserByEmail(c.Request.Context(), loginRequest.Email)
		if err != nil {
			uc.logger.Printf("Failed to find user, error: %v", err)
			// Prevent leaking information about whether the email exists or not
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Verify the provided password with the stored hashed password
		valid, err := VerifyPassword(loginRequest.Password, user.Password)
		if !valid || err != nil {
			uc.logger.Printf("Invalid login attempt for user: %s", loginRequest.Email)
			// Prevent leaking information about whether the email exists or not
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Generate a token (e.g., JWT) for the user (assuming GenerateToken method exists)
		token, refreshToken, err := helper.GenerateAllTokens(user.Email, string(user.User_type), user.ID.Hex())
		if err != nil {
			uc.logger.Printf("Failed to generate token, error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Return token in response to avoid unnecessary allocations or large objects
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   token,
			"refresh": refreshToken,
			"userId":  user.ID.Hex(),
		})
	}
}

func (uc *UserController) GetUserId() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		if userID == "" {
			uc.logger.Println("User ID is missing in the request")
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		user, err := uc.userService.GetUserById(c.Request.Context(), userID)
		if err != nil {
			uc.logger.Printf("Failed to retrieve user, error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func (uc *UserController) EditUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		if userID == "" {
			uc.logger.Println("User ID is missing in the request")
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		var userUpdates user.User
		if err := c.BindJSON(&userUpdates); err != nil {
			uc.logger.Printf("Failed to bind request body, error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		err := uc.userService.UpdateUser(c.Request.Context(), userID, &userUpdates)
		if err != nil {
			uc.logger.Printf("Failed to update user, error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}

func (uc *UserController) GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := uc.userService.GetAllUsers(c.Request.Context())
		if err != nil {
			uc.logger.Printf("Failed to retrieve users, error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

func (uc *UserController) UpdateUserType() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")
		if userID == "" {
			uc.logger.Println("User ID is missing in the request")
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
			return
		}

		var userTypeUpdate struct {
			UserType string `json:"user_type" validate:"required"`
		}

		if err := c.BindJSON(&userTypeUpdate); err != nil {
			uc.logger.Printf("Failed to bind request body, error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		err := uc.userService.UpdateUserType(c.Request.Context(), userID, userTypeUpdate.UserType)
		if err != nil {
			uc.logger.Printf("Failed to update user type, error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user type"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User type updated successfully"})
	}
}

func VerifyPassword(userPassword string, providedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	if err != nil {
		return false, fmt.Errorf("login or password is incorrect")
	}
	return true, nil
}
