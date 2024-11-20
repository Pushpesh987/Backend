package users

import (
	"bytes"
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
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

// UploadProfilePhoto handles the upload of a profile photo and updates the user record.
func UploadProfilePhoto(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(string)

	// Parse the uploaded file
	file, err := c.FormFile("profile_photo")
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "File upload failed", err)
	}

	// Open the file
	fileContent, err := file.Open()
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open file", err)
	}
	defer fileContent.Close()

	// Generate a unique file name
	fileName := uuid.New().String() + "-" + file.Filename

	// Construct the folder structure inside the bucket
	filePath := fmt.Sprintf("profile-photos/%s", fileName)

	// Upload the file to Supabase storage
	publicURL, err := uploadToSupabase(filePath, fileContent)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload file to storage", err)
	}

	// Update user's profile photo URL in DB
	if result := db.Model(&models.User{}).Where("id = ?", userID).Update("profile_photo_url", publicURL); result.Error != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update profile photo URL", result.Error)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Profile photo updated successfully", fiber.Map{"profile_photo_url": publicURL})
}

// uploadToSupabase uploads a file to Supabase storage and returns its public URL.
func uploadToSupabase(fileName string, fileContent io.Reader) (string, error) {
    bucketName := "file-buckets"
    apiURL := os.Getenv("STORAGE_URL") // Should now be https://iczixyjklnvkhqamqaky.supabase.co/storage/v1
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

    // Log the request URL for debugging
    fmt.Println("Request URL:", requestURL)

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





