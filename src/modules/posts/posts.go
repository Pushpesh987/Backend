package posts

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreatePost(c *fiber.Ctx) error {
	db := database.DB
	authID := c.Locals("user_id").(string) // Extract auth_id from JWT

	// Fetch the user's primary key (id) from the users table using auth_id
	var user struct {
		ID string `gorm:"column:id"`
	}
	if err := db.Table("users").Where("auth_id = ?", authID).Select("id").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.HandleError(c, fiber.StatusNotFound, "User not found", err)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch user", err)
	}

	// Convert the string user.ID to uuid.UUID
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Invalid user ID format", err)
	}

	// Parse request body for post details
	body := new(models.Post)
	if err := c.BodyParser(body); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	// Handle media upload if present
	var mediaURL string
	if media, err := c.FormFile("media"); err == nil {
		// Open the file for reading
		mediaContent, err := media.Open()
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open media file", err)
		}
		defer mediaContent.Close()

		// Generate a unique file name
		fileName := uuid.New().String() + "-" + media.Filename

		// Upload the media file to Supabase
		mediaURL, err = uploadToSupabase(fileName, mediaContent)
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload media", err)
		}
	}

	// Create the post object with the necessary fields
	post := models.Post{
		UserID:  userID,   // Use the parsed userID (uuid.UUID)
		Content: body.Content,
		MediaURL: mediaURL,  // Media URL, if any
	}

	// Insert the new post into the database
	if err := db.Table("posts").Create(&post).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create post", err)
	}

	// Respond with the created post details
	return helpers.HandleSuccess(c, fiber.StatusOK, "Post created successfully", post)
}

// Reusing the existing uploadToSupabase function here
func uploadToSupabase(fileName string, fileContent io.Reader) (string, error) {
	bucketName := "file-buckets"
	apiURL := os.Getenv("STORAGE_URL") // Example: https://iczixyjklnvkhqamqaky.supabase.co/storage/v1
	authToken := "Bearer " + os.Getenv("SERVICE_ROLE_SECRET")

	if apiURL == "" {
		return "", fmt.Errorf("STORAGE_URL is not set in the environment variables")
	}

	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf("failed to create multipart file: %w", err)
	}
	_, err = io.Copy(part, fileContent)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}
	writer.Close()

	// Construct REST API URL for storage
	requestURL := fmt.Sprintf("%s/object/%s/%s", apiURL, bucketName, fileName)

	// Make the HTTP request
	req, err := http.NewRequest("POST", requestURL, body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if upload succeeded
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println("Upload failed. Response Body:", string(respBody))
		return "", fmt.Errorf("upload failed with status: %s", resp.Status)
	}

	// Construct the public URL
	publicURL := fmt.Sprintf("%s/object/public/%s/%s", apiURL, bucketName, fileName)
	return publicURL, nil
}
