package communities

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"strconv"

	// "bytes"
	"fmt"
	"log"

	// "io"
	// "mime/multipart"
	// "net/http"
	// "os"
	// "strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	// "gorm.io/gorm"
)

func CreateCommunity(c *fiber.Ctx) error {
	db := database.DB

	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		log.Println("Invalid or missing userID")
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing auth_id", nil)
	}

	userID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Error parsing user ID as UUID: %v\n", err)
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}
	fmt.Println("Retrieved userID:", userID)

	body := new(models.Community)
	if err := c.BodyParser(body); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	body.CreatedAt = time.Now()

	if result := db.Create(&body); result.Error != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create community", result.Error)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Community created successfully", body)
}

func JoinCommunity(c *fiber.Ctx) error {
	db := database.DB

	// Get user_id from JWT locals
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
	}

	// Parse userID string to uuid.UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	// Get community ID from URL params and convert it to int
	communityIDStr := c.Params("id")
	communityID, err := strconv.Atoi(communityIDStr)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid community ID format", err)
	}

	var communityMember models.CommunityMember

	// Check if the user is already a member of the community
	if err := db.Where("user_id = ? AND community_id = ?", userUUID, communityID).First(&communityMember).Error; err == nil {
		return helpers.HandleError(c, fiber.StatusConflict, "User is already a member", nil)
	}

	// Create new membership record
	communityMember.UserID = userUUID
	communityMember.CommunityID = communityID
	if err := db.Create(&communityMember).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to join community", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Successfully joined the community", communityMember)
}

func GetCommunityDetails(c *fiber.Ctx) error {
	db := database.DB
	communityID := c.Params("id")

	var community models.Community
	if err := db.First(&community, communityID).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusNotFound, "Community not found", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Community details fetched successfully", community)
}

func GetAllCommunities(c *fiber.Ctx) error {
	db := database.DB

	var communities []models.Community
	if err := db.Find(&communities).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch communities", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Communities fetched successfully", communities)
}

func LeaveCommunity(c *fiber.Ctx) error {
	db := database.DB
	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		log.Println("Invalid or missing userID")
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing auth_id", nil)
	}

	userID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Error parsing user ID as UUID: %v\n", err)
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	communityID := c.Params("id")
	var communityMember models.CommunityMember

	if err := db.Where("user_id = ? AND community_id = ?", userID, communityID).First(&communityMember).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusNotFound, "User not a member of the community", err)
	}

	if err := db.Delete(&communityMember).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to leave the community", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Successfully left the community", nil)
}

func GetUserCommunities(c *fiber.Ctx) error {
    db := database.DB
    userID, ok := c.Locals("user_id").(string)
    if !ok || userID == "" {
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
    }

    var communities []models.Community
    if err := db.Joins("JOIN community_members ON communities.id = community_members.community_id").
        Where("community_members.user_id = ?", userID).
        Find(&communities).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch user communities", err)
    }

    return helpers.HandleSuccess(c, fiber.StatusOK, "User communities fetched successfully", communities)
}

// func GetCommunityMessages(c *fiber.Ctx) error {
// 	db := database.DB
// 	communityID := c.Params("id")

// 	var messages []models.Message
// 	if err := db.Where("community_id = ?", communityID).Find(&messages).Error; err != nil {
// 		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch messages", err)
// 	}

// 	return helpers.HandleSuccess(c, fiber.StatusOK, "Messages fetched successfully", messages)
// }

func GetCommunityMessages(c *fiber.Ctx) error {
    db := database.DB
    communityID := c.Params("id")

    type MessageWithUser struct {
		ID          int       `json:"id"`
		CommunityID int       `json:"community_id"`
		UserID      uuid.UUID `json:"user_id"`
		Username    string    `json:"username"`
		Message     string    `json:"message"`
		CreatedAt   time.Time `json:"created_at"`
	}

    var messages []MessageWithUser
    query := `
        SELECT m.id AS message_id, m.user_id, u.username, m.message, m.created_at
        FROM messages m
        JOIN users u ON m.user_id = u.id
        WHERE m.community_id = ?
        ORDER BY m.created_at DESC
    `
    if err := db.Raw(query, communityID).Scan(&messages).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch messages", err)
    }

    return helpers.HandleSuccess(c, fiber.StatusOK, "Messages fetched successfully", messages)
}
