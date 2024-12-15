package iotlogs

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"Backend/src/modules/notifications"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateIotLog(c *fiber.Ctx) error {
	db := database.DB

	// Parse the request body to create a new IoT log entry
	body := new(models.IotLog)
	if err := c.BodyParser(body); err != nil {
		log.Printf("Error parsing request body: %v\n", err)
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	// Validate the user_id
	if body.UserID == uuid.Nil {
		log.Println("Missing or invalid user_id in request body")
		return helpers.HandleError(c, fiber.StatusBadRequest, "Missing or invalid user_id", nil)
	}

	// Initialize the timestamp
	body.Timestamp = time.Now()

	// Create the IoT log in the database
	if result := db.Create(&body); result.Error != nil {
		log.Printf("Error creating IoT log: %v\n", result.Error)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create IoT log", result.Error)
	}

	// Trigger the notification logic
	go sendConnectionNotifications(db, body)

	return helpers.HandleSuccess(c, fiber.StatusCreated, "IoT log created successfully", body)
}

// sendConnectionNotifications handles finding nearby IoT logs, matching interests, and sending notifications
func sendConnectionNotifications(db *gorm.DB, currentLog *models.IotLog) {
    // Define the time window for recent IoT logs (15-20 minutes)
    timeWindow := time.Now().Add(-20 * time.Minute)

    // Fetch nearby IoT logs within the time window and matching location
    var nearbyLogs []models.IotLog
    db.Where("timestamp >= ? AND location = ?", timeWindow, currentLog.Location).Find(&nearbyLogs)

    // Fetch the current user's interests by joining user_interests and interests tables
    var user1InterestIDs []uuid.UUID
    db.Table("user_interests").Where("user_id = ?", currentLog.UserID).Pluck("interest_id", &user1InterestIDs)

    // Fetch the actual interest models for the current user
    var user1InterestModels []models.Interest
    db.Where("interest_id IN ?", user1InterestIDs).Find(&user1InterestModels)

    // Loop through nearby logs to check for matches and send notifications
    for _, log := range nearbyLogs {
        // Skip the current log entry (don't notify the same user)
        if log.UserID == currentLog.UserID {
            continue
        }

        // Fetch the other user's interests by joining user_interests and interests tables
        var user2InterestIDs []uuid.UUID
        db.Table("user_interests").Where("user_id = ?", log.UserID).Pluck("interest_id", &user2InterestIDs)

        // Fetch the actual interest models for the other user
        var user2InterestModels []models.Interest
        db.Where("interest_id IN ?", user2InterestIDs).Find(&user2InterestModels)

        // Check if there's any matching interest
        if hasMatchingInterest(user1InterestModels, user2InterestModels) {
            // Ensure that no recent notifications have been sent
            var recentNotifications []models.Notification
            db.Where("user_id IN ? AND created_at >= ?", []uuid.UUID{currentLog.UserID, log.UserID}, time.Now().Add(-4*24*time.Hour)).Find(&recentNotifications)

            // If no recent notifications, send the new ones
            if len(recentNotifications) == 0 {
                // Select a random notification template
                var templates []models.NotificationTemplate
                db.Where("category = ?", "connection").Find(&templates)
                rand.Seed(time.Now().UnixNano())
                selectedTemplate := templates[rand.Intn(len(templates))]

                // Replace placeholders with actual usernames
                var user1 models.User
                db.Where("id = ?", currentLog.UserID).First(&user1)
                var user2 models.User
                db.Where("id = ?", log.UserID).First(&user2)

                message1 := strings.Replace(selectedTemplate.TemplateText, "{user1}", user1.Username, -1)
                message1 = strings.Replace(message1, "{user2}", user2.Username, -1)

                message2 := strings.Replace(selectedTemplate.TemplateText, "{user1}", user2.Username, -1)
                message2 = strings.Replace(message2, "{user2}", user1.Username, -1)

                // Create and save the notifications
                notification1 := models.Notification{
                    UserID:    currentLog.UserID,
                    Message:   message1,
                    Category:  "connection",
                    CreatedAt: time.Now(),
                }
                notification2 := models.Notification{
                    UserID:    log.UserID,
                    Message:   message2,
                    Category:  "connection",
                    CreatedAt: time.Now(),
                }

                db.Create(&notification1)
                db.Create(&notification2)

                // Trigger real-time notifications using WebSockets
                go notifications.SendNotification(currentLog.UserID.String(), "", message1)
                go notifications.SendNotification(log.UserID.String(), "", message2)
            }
        }
    }
}


// Helper function to check if there's a matching interest between user1 and user2
func hasMatchingInterest(user1Interests, user2Interests []models.Interest) bool {
    for _, u1 := range user1Interests {
        for _, u2 := range user2Interests {
            if u1.InterestName == u2.InterestName { // Updated field name
                return true
            }
        }
    }
    return false
}



