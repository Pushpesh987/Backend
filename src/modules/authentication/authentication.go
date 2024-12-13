package authentication

import (
	"Backend/src/core/config"
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"fmt"
	"log"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func issueJwtToken(authID string, userID string, email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["sub"] = authID
	claims["user_id"] = userID
	claims["email"] = email
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(30 * 24 * time.Hour).Unix()

	secretKey := config.Config("JWT_SECRET")
	return token.SignedString([]byte(secretKey))
}

func SignUp(c *fiber.Ctx) error {
	db := database.DB
	if db == nil {
		log.Fatal("Database connection is not initialized")
	}

	auth := new(models.Auth)
	if err := c.BodyParser(auth); err != nil {
		log.Printf("Error parsing body: %v\n", err)
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	if auth.Email == "" || auth.Password == "" || auth.Username == "" {
		log.Println("Missing required fields: email, username, or password")
		return helpers.HandleError(c, fiber.StatusBadRequest, "Email, username, and password are required", nil)
	}

	var existingAuth models.Auth
	if err := db.Where("email = ?", auth.Email).Or("username = ?", auth.Username).First(&existingAuth).Error; err == nil {
		log.Println("Email or username already exists")
		return helpers.HandleError(c, fiber.StatusConflict, "Email or username already exists", nil)
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(auth.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to hash password", err)
	}
	auth.ID = uuid.New()
	auth.Password = string(hashedPwd)

	if result := db.Create(auth); result.Error != nil {
		log.Printf("Error creating auth record: %v\n", result.Error)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create account", result.Error)
	}

	user := models.User{
		ID:        uuid.New(),
		AuthID:    auth.ID,
		Username:  auth.Username,
		Email:     auth.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if result := db.Create(&user); result.Error != nil {
		db.Delete(auth)
		log.Printf("Error creating user record: %v\n", result.Error)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create user record", result.Error)
	}

	userBadge := models.UserBadge{
		UserID:  user.ID,
		BadgeID: 1,
	}
	
	if err := db.Create(&userBadge).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to assign badge to user", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Account created successfully", map[string]interface{}{
		"auth_id": auth.ID,
		"user_id": user.ID,
	})
}

func SignIn(c *fiber.Ctx) error {
	db := database.DB
	auth := new(models.Auth)
	fetchedUser := new(models.Auth)

	if err := c.BodyParser(auth); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	if auth.Email == "" || auth.Password == "" {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Email and password are required", nil)
	}

	fmt.Println("Attempting to fetch user by email:", auth.Email)

	result := db.Where("email = ?", auth.Email).First(&fetchedUser)
	if result.Error != nil {
		fmt.Println("Error fetching user:", result.Error)
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid email credentials", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(fetchedUser.Password), []byte(auth.Password)); err != nil {
		fmt.Println("Password mismatch:", err)
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid password credentials", nil)
	}

	user := new(models.User)
	if err := db.Where("auth_id = ?", fetchedUser.ID).First(&user).Error; err != nil {
		fmt.Println("Error fetching user:", err)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch user details", err)
	}

	token, err := issueJwtToken(fetchedUser.ID.String(),
	 user.ID.String(), 
	 fetchedUser.Email)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to generate token", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Sign-in successful", fiber.Map{"token": token})
}







