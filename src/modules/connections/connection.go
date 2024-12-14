package connection

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func Follow(c *fiber.Ctx) error {
	db := database.DB

	var input struct {
		ConnectionID string `json:"connection_id"`
	}
	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
	}
	userID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Error parsing user ID as UUID: %v\n", err)
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}
	fmt.Println("Retrieved userID:", userID)

	if err := c.BodyParser(&input); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	parsedConnectionID, err := uuid.Parse(input.ConnectionID)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid connection_id format", err)
	}

	connection := models.Connection{
		UserID:       userID,
		ConnectionID: parsedConnectionID,
	}

	if err := db.Create(&connection).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create connection", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Successfully followed the user", connection)
}

func ConnectionCheck(c *fiber.Ctx) error {
	db := database.DB

	var input struct {
		ConnectionID string `json:"connection_id"`
	}

	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
	}
	userID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Error parsing user ID as UUID: %v\n", err)
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}
	fmt.Println("Retrieved userID:", userID)

	if err := c.BodyParser(&input); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	if input.ConnectionID == "" {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Missing connection_id", nil)
	}

	parsedConnectionID, err := uuid.Parse(input.ConnectionID)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid connection_id format", err)
	}

	var viewerToOther models.Connection
	var otherToViewer models.Connection

	viewerFollowing := db.Where("user_id = ? AND connection_id = ?", userID, parsedConnectionID).First(&viewerToOther).Error == nil
	otherFollowing := db.Where("user_id = ? AND connection_id = ?", parsedConnectionID, userID).First(&otherToViewer).Error == nil

	var status string
	if viewerFollowing && otherFollowing {
		status = "Following"
	} else if otherFollowing {
		status = "Follow Back"
	} else {
		status = "Follow"
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Connection status retrieved", map[string]string{
		"status": status,
	})
}
