package posts

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreatePost(c *fiber.Ctx) error {
	db := database.DB

	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		log.Println("Invalid or missing userId")
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
	}

	userID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Error parsing user ID as UUID: %v\n", err)
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	content := c.FormValue("content")
	if content == "" {
		log.Println("Post content is empty")
		return helpers.HandleError(c, fiber.StatusBadRequest, "Post content cannot be empty", nil)
	}

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

	post := models.Post{
		UserID:        userID,
		Content:       content,
		MediaURL:      mediaURL,
		LikesCount:    0,
		CommentsCount: 0,
	}
	if err := db.Table("posts").Create(&post).Error; err != nil {
		log.Printf("Error creating post: %v\n", err)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create post", err)
	}

	var postTags []models.PostTag
	for _, tag := range tags {
		var tagID int

		err := db.Table("tags").Where("tag = ?", tag).Select("id").Scan(&tagID).Error
		if err != nil {
			log.Printf("Error finding tag ID for tag %s: %v\n", tag, err)
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to find tag ID", err)
		}

		if tagID == 0 {
			err := db.Table("tags").Create(&models.Tag{Tag: tag}).Error
			if err != nil {
				log.Printf("Error inserting tag %s: %v\n", tag, err)
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to insert tag", err)
			}

			err = db.Table("tags").Where("tag = ?", tag).Select("id").Scan(&tagID).Error
			if err != nil {
				log.Printf("Error finding new tag ID for tag %s: %v\n", tag, err)
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to find new tag ID", err)
			}
		}

		var exists bool
		err = db.Table("post_tags").
			Select("exists (select 1 from post_tags where post_id = ? and tag_id = ?)", post.ID, tagID).
			Scan(&exists).Error
		if err != nil {
			log.Printf("Error checking for duplicate tag entry: %v\n", err)
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Error checking for duplicate tag entry", err)
		}

		if !exists {

			postTags = append(postTags, models.PostTag{
				PostID: post.ID,
				TagID:  tagID,
			})
		} else {
			log.Printf("Post tag entry already exists for post %s and tag %d\n", post.ID, tagID)
		}
	}

	if len(postTags) > 0 {
		if err := db.Table("post_tags").Create(&postTags).Error; err != nil {
			log.Printf("Error creating post tags: %v\n", err)
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create post tags", err)
		}
	}

	log.Printf("Post created successfully: %+v\n", post)
	return helpers.HandleSuccess(c, fiber.StatusOK, "Post created successfully", post)
}

func getPredictedTags(content string) ([]string, error) {
	if content == "" {
		log.Println("Content is empty; cannot predict tags.")
		return nil, fmt.Errorf("content cannot be empty for tag prediction")
	}
	modelURL := "https://ml-models-1rr1.onrender.com/predict"
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
	folderName := "posts"
	apiURL := os.Getenv("STORAGE_URL")
	authToken := "Bearer " + os.Getenv("SERVICE_ROLE_SECRET")

	if apiURL == "" {
		return "", fmt.Errorf("STORAGE_URL is not set in the environment variables")
	}

	// Add timestamp to filename
	timestamp := time.Now().Unix()
	uniqueFileName := fmt.Sprintf("%d_%s", timestamp, fileName)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", uniqueFileName)
	if err != nil {
		return "", fmt.Errorf("failed to create multipart file: %w", err)
	}
	_, err = io.Copy(part, fileContent)
	if err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}
	writer.Close()

	objectPath := fmt.Sprintf("%s/%s", folderName, uniqueFileName)
	requestURL := fmt.Sprintf("%s/object/%s/%s", apiURL, bucketName, objectPath)

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

	publicURL := fmt.Sprintf("%s/object/public/%s/%s", apiURL, bucketName, objectPath)
	return publicURL, nil
}


func CreateComment(c *fiber.Ctx) error {
	db := database.DB

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Unauthorized: missing user_id", nil)
	}

	type Request struct {
		PostID  string `json:"post_id" validate:"required,uuid"`
		Content string `json:"content" validate:"required"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	var postExists bool
	if err := db.Raw("SELECT EXISTS (SELECT 1 FROM posts WHERE id = ?)", req.PostID).Scan(&postExists).Error; err != nil || !postExists {
		return helpers.HandleError(c, fiber.StatusNotFound, "Post not found", err)
	}

	comment := models.Comment{
		ID:      uuid.New(),
		UserID:  uuid.MustParse(userID),
		PostID:  uuid.MustParse(req.PostID),
		Content: req.Content,
	}
	if err := db.Create(&comment).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create comment", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Comment created successfully", comment)
}

func CreateLike(c *fiber.Ctx) error {
	db := database.DB

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Unauthorized: missing user_id", nil)
	}

	type Request struct {
		PostID string `json:"post_id" validate:"required,uuid"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	var postExists bool
	if err := db.Raw("SELECT EXISTS (SELECT 1 FROM posts WHERE id = ?)", req.PostID).Scan(&postExists).Error; err != nil || !postExists {
		return helpers.HandleError(c, fiber.StatusNotFound, "Post not found", err)
	}

	var existingLike models.Like
	if err := db.Where("user_id = ? AND post_id = ?", userID, req.PostID).First(&existingLike).Error; err == nil {
		if err := db.Where("user_id = ? AND post_id = ?", userID, req.PostID).Delete(&models.Like{}).Error; err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to remove like", err)
		}
		return helpers.HandleSuccess(c, fiber.StatusOK, "Like removed successfully", nil)
	}

	like := models.Like{
		UserID: uuid.MustParse(userID),
		PostID: uuid.MustParse(req.PostID),
	}
	if err := db.Create(&like).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create like", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Like created successfully", nil)
}


func CreateShare(c *fiber.Ctx) error {
	db := database.DB

	type Request struct {
		PostID   string `json:"post_id" validate:"required,uuid"`
		ToUserID string `json:"to_user_id" validate:"required,uuid"`
	}
	var req Request
	if err := c.BodyParser(&req); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid request payload", err)
	}

	var exists bool
	if err := db.Raw("SELECT EXISTS (SELECT 1 FROM posts WHERE id = ?)", req.PostID).Scan(&exists).Error; err != nil || !exists {
		return helpers.HandleError(c, fiber.StatusNotFound, "Post not found", err)
	}

	userID := c.Locals("user_id")
	if userID == nil {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "User authentication failed: user_id is missing", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Invalid user_id type", nil)
	}
	userIDParsed, err := uuid.Parse(userIDStr)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Invalid user_id format", err)
	}

	fmt.Println("Auth ID:", userIDParsed)

	var user models.User
	if err := db.First(&user, "id = ?", userIDParsed).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "User not found", err)
	}

	fmt.Println("Fetched user details:", user)

	share := models.Share{
		ID:         uuid.New(),
		FromUserID: user.ID,
		ToUserID:   uuid.MustParse(req.ToUserID),
		PostID:     uuid.MustParse(req.PostID),
	}
	if err := db.Create(&share).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to share post", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Post shared successfully", share)
}

func GetLikesCount(c *fiber.Ctx) error {
	db := database.DB

	postID := c.Params("post_id")
	if postID == "" {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Missing post ID", nil)
	}

	var postExists bool
	if err := db.Raw("SELECT EXISTS (SELECT 1 FROM posts WHERE id = ?)", postID).Scan(&postExists).Error; err != nil || !postExists {
		return helpers.HandleError(c, fiber.StatusNotFound, "Post not found", err)
	}

	var likesCount int64
	if err := db.Raw("SELECT COUNT(*) FROM likes WHERE post_id = ?", postID).Scan(&likesCount).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to retrieve likes count", err)
	}

	response := map[string]interface{}{
		"post_id":     postID,
		"likes_count": likesCount,
	}
	return helpers.HandleSuccess(c, fiber.StatusOK, "Likes count retrieved successfully", response)
}
