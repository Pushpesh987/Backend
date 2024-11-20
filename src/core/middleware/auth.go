package middleware

import (
	"Backend/src/core/config"  // Adjust this import to your config package path
	"Backend/src/core/helpers" // Adjust this import to your helpers package path

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected middleware for validating JWT tokens
func Protected() fiber.Handler {
	jwtSecret := config.Config("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET is not set in the environment variables") // Panic to prevent startup
	}

	return jwtware.New(jwtware.Config{
		SigningKey:   jwtware.SigningKey{Key: []byte(jwtSecret)},
		ErrorHandler: jwtError,
		SuccessHandler: func(c *fiber.Ctx) error {
			// Extract user claims and attach user_id to the context
			user := c.Locals("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)
			if userID, ok := claims["sub"].(string); ok {
				c.Locals("user_id", userID)
				return c.Next()
			}
			return helpers.HandleError(c, fiber.StatusUnauthorized, "User ID missing in token", nil)
		},
	})
}

// jwtError handles JWT-related errors
func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Missing or malformed JWT", err)
	}
	return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or expired JWT", err)
}
