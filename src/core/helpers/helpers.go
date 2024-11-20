package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Initialize a validator instance using go-playground's validator package
var Validator = validator.New()

// Validate checks the struct fields against the specified validation tags.
func Validate(val interface{}) error {
	return Validator.Struct(val)
}

// HandleSuccess sends a structured JSON response for successful requests.
func HandleSuccess(context *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return context.Status(statusCode).JSON(fiber.Map{
		"status":  "success",
		"message": message,
		"error":   nil,
		"data":    data,
	})
}

// HandleError sends a structured JSON response for errors.
func HandleError(context *fiber.Ctx, statusCode int, message string, err error) error {
	return context.Status(statusCode).JSON(fiber.Map{
		"status":  "error",
		"message": message,
		"error":   err.Error(),
		"data":    nil,
	})
}

// GenerateErrorResponse creates a custom Fiber-compatible error response object.
func GenerateErrorResponse(message string, err error) fiber.Map {
	return fiber.Map{
		"status":  "error",
		"message": message,
		"error":   err.Error(),
		"data":    nil,
	}
}
