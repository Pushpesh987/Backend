package users

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"

	"github.com/gofiber/fiber/v2"
)

// GetUserDetails retrieves the profile of the authenticated user.
func GetUserDetails(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(string)

	// Fetch user from DB
	user := new(models.User)
	if result := db.First(user, "id = ?", userID); result.Error != nil {
		return helpers.HandleError(c, fiber.StatusNotFound, "User not found", result.Error)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "User details retrieved successfully", user)
}

// UpdateUserDetails updates the profile of the authenticated user.
func UpdateUserDetails(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(string)

	// Parse request body
	body := new(models.User)
	if err := c.BodyParser(body); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	// Update user details in DB
	if result := db.Model(&models.User{}).Where("id = ?", userID).Updates(body); result.Error != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user details", result.Error)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "User details updated successfully", nil)
}
