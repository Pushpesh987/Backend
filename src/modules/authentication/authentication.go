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
func issueJwtToken(userID string, firstName string, lastName string, email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["sub"] = userID
	claims["first_name"] = firstName
	claims["last_name"] = lastName
	claims["email"] = email
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(30 * 24 * time.Hour).Unix()

	secretKey := config.Config("JWT_SECRET")
	return token.SignedString([]byte(secretKey))
}

// SignUp handles user registration.
func SignUp(c *fiber.Ctx) error {
	db := database.DB
	auth := new(models.Auth)

	// Parse request body for auth details
	if err := c.BodyParser(auth); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	// Validate required fields
	if auth.Email == "" || auth.Password == "" || auth.Username == "" {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Email, username, and password are required", nil)
	}

	// Check for duplicate email or username
	var existingAuth models.Auth
	if err := db.Where("email = ?", auth.Email).Or("username = ?", auth.Username).First(&existingAuth).Error; err == nil {
		return helpers.HandleError(c, fiber.StatusConflict, "Email or username already exists", nil)
	}

	// Hash the password using bcrypt
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(auth.Password), bcrypt.DefaultCost)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to hash password", err)
	}
	auth.ID = uuid.New()
	auth.Password = string(hashedPwd)

	// Save the Auth record to the database
	if result := db.Create(auth); result.Error != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create account", result.Error)
	}

	// Create a minimal User record
	user := models.User{
		ID:        uuid.New(),
		AuthID:    auth.ID,
		Username:  auth.Username,
		Email:     auth.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if result := db.Create(&user); result.Error != nil {
		// Roll back auth record if user creation fails
		db.Delete(auth)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create user record", result.Error)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Account created successfully", map[string]interface{}{
		"auth_id": auth.ID,
		"user_id": user.ID,
	})
}


// SignIn handles user authentication.
func SignIn(c *fiber.Ctx) error {
    db := database.DB
    auth := new(models.Auth)
    fetchedUser := new(models.Auth)

    // Parse request body to get email and password
    if err := c.BodyParser(auth); err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
    }

    // Debug: Log the email being fetched
    fmt.Println("Attempting to fetch user by email:", auth.Email)

    // Fetch user by email
    result := db.Where("email = ?", auth.Email).First(&fetchedUser)
    if result.Error != nil {
        fmt.Println("Error fetching user:", result.Error) // Debug error
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid email credentials", result.Error)
    }

    // Debug: Print password details
    fmt.Println("Plain password entered:", auth.Password)
    fmt.Println("Stored hashed password:", fetchedUser.Password)

    // Compare provided password with hashed password
    if err := bcrypt.CompareHashAndPassword([]byte(fetchedUser.Password), []byte(auth.Password)); err != nil {
        fmt.Println("Password mismatch:", err) // Debug error
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid password credentials", err)
    }

    // Generate JWT token on successful authentication
    token, err := issueJwtToken(fetchedUser.ID.String(), "", "", fetchedUser.Email)
    if err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to generate token", err)
    }

    // Return success response with the token
    return helpers.HandleSuccess(c, fiber.StatusOK, "Sign-in successful", fiber.Map{"token": token})
}







