package posts

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TextArray []string

func (ta *TextArray) Scan(value interface{}) error {
	if value == nil {
		*ta = []string{}
		return nil
	}
	arrayStr := string(value.([]byte))
	arrayStr = strings.Trim(arrayStr, "{}")
	if arrayStr == "" {
		*ta = []string{}
		return nil
	}
	*ta = strings.Split(arrayStr, ",")
	return nil
}

func (ta TextArray) Value() (driver.Value, error) {
	// Format as PostgreSQL array
	return "{" + strings.Join(ta, ",") + "}", nil
}

func formatTagsForPostgres(tags []string) string {
	return fmt.Sprintf("{%s}", strings.Join(tags, ","))
}

func CreatePost(c *fiber.Ctx) error {
	db := database.DB

	authID, ok := c.Locals("user_id").(string)
	if !ok || authID == "" {
		log.Println("Invalid or missing authID")
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing auth_id", nil)
	}
	log.Printf("authID from JWT: %s\n", authID)

	var user struct {
		ID string `gorm:"column:id"`
	}
	if err := db.Table("users").Where("auth_id = ?", authID).Select("id").First(&user).Error; err != nil {
		log.Printf("Error fetching user: %v\n", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.HandleError(c, fiber.StatusNotFound, "User not found", nil)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Database query failed", err)
	}
	log.Printf("Fetched user.ID: %s\n", user.ID)

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		log.Printf("Error parsing user ID as UUID: %v\n", err)
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	content := c.FormValue("content")
	if content == "" {
		log.Println("Post content is empty")
		return helpers.HandleError(c, fiber.StatusBadRequest, "Post content cannot be empty", nil)
	}
	log.Printf("Parsed content: %s\n", content)

	var mediaURL string
	media, err := c.FormFile("media")
	if err == nil {
		mediaContent, err := media.Open()
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open media file", err)
		}
		defer mediaContent.Close()

		fileName := uuid.New().String() + "-" + media.Filename
		mediaURL, err = uploadToSupabase(fileName, mediaContent)
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload media", err)
		}
	} else if err != http.ErrMissingFile {
		log.Printf("Media upload error: %v\n", err)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Unexpected media upload error", err)
	}

	tags, err := getPredictedTags(content)
	if err != nil {
		log.Printf("Error predicting tags: %v\n", err)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to predict tags", err)
	}

	postgresArray := formatTagsForPostgres(tags) 

	post := models.Post{
		UserID:        userID,
		Content:       content,
		MediaURL:      mediaURL,
		Tags:          postgresArray,
		LikesCount:    0,
		CommentsCount: 0,
	}

	if err := db.Table("posts").Create(&post).Error; err != nil {
		log.Printf("Error creating post: %v\n", err)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create post", err)
	}

	log.Printf("Post created successfully: %+v\n", post)
	return helpers.HandleSuccess(c, fiber.StatusOK, "Post created successfully", post)
}

func getPredictedTags(content string) ([]string, error) {
	if content == "" {
		log.Println("Content is empty; cannot predict tags.")
		return nil, fmt.Errorf("content cannot be empty for tag prediction")
	}
	modelURL := "http://localhost:5000/predict" // Update with actual endpoint
	requestData := map[string]string{"content": content}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		log.Printf("Error marshaling request data: %v\n", err)
		return nil, err
	}

	resp, err := http.Post(modelURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Error calling tag prediction model: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Tag prediction model responded with status: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch predicted tags")
	}

	var response struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Error decoding model response: %v\n", err)
		return nil, err
	}

	log.Printf("Predicted tags: %v\n", response.Tags)
	return response.Tags, nil
}

func uploadToSupabase(fileName string, fileContent io.Reader) (string, error) {
	bucketName := "file-buckets"
	apiURL := os.Getenv("STORAGE_URL") 
	authToken := "Bearer " + os.Getenv("SERVICE_ROLE_SECRET")

	if apiURL == "" {
		return "", fmt.Errorf("STORAGE_URL is not set in the environment variables")
	}

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

	requestURL := fmt.Sprintf("%s/object/%s/%s", apiURL, bucketName, fileName)

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

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println("Upload failed. Response Body:", string(respBody))
		return "", fmt.Errorf("upload failed with status: %s", resp.Status)
	}

	publicURL := fmt.Sprintf("%s/object/public/%s/%s", apiURL, bucketName, fileName)
	return publicURL, nil
}
