package connection

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Follow API
func Follow(c *fiber.Ctx) error {
    db := database.DB

    var input struct {
        ConnectionID string `json:"connection_id"`
    }

    // Extract auth_id from JWT
    authID, ok := c.Locals("user_id").(string)
    if !ok || authID == "" {
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing auth_id", nil)
    }

    // Fetch user_id from the database using auth_id
    var user struct {
        ID uuid.UUID `gorm:"column:id"`
    }
    if err := db.Table("users").Select("id").Where("auth_id = ?", authID).First(&user).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch user_id", err)
    }

    if err := c.BodyParser(&input); err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
    }

    // Parse input.ConnectionID into uuid.UUID
    parsedConnectionID, err := uuid.Parse(input.ConnectionID)
    if err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid connection_id format", err)
    }

    // Insert a new connection
    connection := models.Connection{
        UserID:       user.ID,
        ConnectionID: parsedConnectionID,
    }

    if err := db.Create(&connection).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create connection", err)
    }

    return helpers.HandleSuccess(c, fiber.StatusCreated, "Successfully followed the user", connection)
}

// ConnectionCheck API
func ConnectionCheck(c *fiber.Ctx) error {
    db := database.DB

    var input struct {
        ConnectionID string `json:"connection_id"`
    }

    // Extract auth_id from JWT
    authID, ok := c.Locals("user_id").(string)
    if !ok || authID == "" {
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing auth_id", nil)
    }

    // Fetch user_id from the database using auth_id
    var viewer struct {
        ID uuid.UUID `gorm:"column:id"`
    }
    if err := db.Table("users").Select("id").Where("auth_id = ?", authID).First(&viewer).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch user_id", err)
    }

    // Parse and validate input
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

    // Check the connection status
    var viewerToOther models.Connection
    var otherToViewer models.Connection

    viewerFollowing := db.Where("user_id = ? AND connection_id = ?", viewer.ID, parsedConnectionID).First(&viewerToOther).Error == nil
    otherFollowing := db.Where("user_id = ? AND connection_id = ?", parsedConnectionID, viewer.ID).First(&otherToViewer).Error == nil

    // Determine the relationship status
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


