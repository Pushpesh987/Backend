package authentication

import (
	"fmt"
	"Backend/src/core/config"
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid" 
	"time"
)

// issueJwtToken generates a JWT token for authenticated users.
func issueJwtToken(userID string, name string, email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["sub"] = userID
	claims["name"] = name
	claims["email"] = email
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(30 * 24 * time.Hour).Unix()

	secretKey := config.Config("JWT_SECRET")
	return token.SignedString([]byte(secretKey))
}

// SignUp handles user registration.
func SignUp(c *fiber.Ctx) error {
	db := database.DB
	body := new(models.User)

	// Parse request body
	if err := c.BodyParser(body); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	// Generate UUID for the user ID
	body.ID = uuid.New()

	// Hash password
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to hash password", err)
	}

	body.Password = string(hashedPwd)

	// Create user in DB
	if result := db.Create(body); result.Error != nil {
		// Log the error to the console for debugging
		fmt.Println("Error creating user:", result.Error)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create user account", result.Error)
	}

	// Generate JWT token
	token, err := issueJwtToken(body.ID.String(), body.FirstName, body.Email)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to generate token", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Account created successfully", fiber.Map{"token": token})
}

// SignIn handles user authentication.
func SignIn(c *fiber.Ctx) error {
	db := database.DB
	body := new(models.User)

	// Parse request body
	if err := c.BodyParser(body); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	// Fetch user from DB
	user := new(models.User)
	if result := db.Where("email = ?", body.Email).First(user); result.Error != nil {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid login credentials", result.Error)
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid login credentials", err)
	}

	// Generate JWT token
	token, err := issueJwtToken(user.ID.String(), user.FirstName, user.Email)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to generate token", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Sign-in successful", fiber.Map{"token": token})
}
