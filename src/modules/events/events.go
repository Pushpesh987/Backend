package events

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"bytes"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func CreateEvent(c *fiber.Ctx) error {
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

	body := new(models.Event)
	body.UserID = userID

	if err := c.BodyParser(body); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Failed to parse form data", err)
	}
	files := form.File["media"]
	var mediaURL string
	if len(files) > 0 {
		mediaFile := files[0]

		file, err := mediaFile.Open()
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open media file", err)
		}
		defer file.Close()

		publicURL, err := uploadToSupabase(mediaFile.Filename, file)
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload media to Supabase", err)
		}

		body.Media = publicURL
		mediaURL = publicURL
	}
	log.Printf("mediaURL: %v", mediaURL)
	if result := db.Create(&body); result.Error != nil {
		if err := deleteFileFromSupabase(body.Media); err != nil {
			log.Printf("Failed to delete media file after event creation failure: %v\n", err)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create event", result.Error)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Event created successfully", body)
}

func CreateWorkshop(c *fiber.Ctx) error {
	db := database.DB

	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		log.Println("Invalid or missing userID")
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
	}

	userID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Error parsing user ID as UUID: %v\n", err)
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}
	fmt.Println("Retrieved userID:", userID)
	body := new(models.Workshop)
	body.UserID = userID

	if err := c.BodyParser(body); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	if body.Duration != "" {
		parsedDuration := body.Duration
		log.Printf("Parsed duration: %v", parsedDuration)
	} else {
		log.Println("Duration not provided")
	}

	form, err := c.MultipartForm()
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Failed to parse form data", err)
	}
	files := form.File["media"]
	var mediaURL string
	if len(files) > 0 {
		mediaFile := files[0]

		file, err := mediaFile.Open()
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open media file", err)
		}
		defer file.Close()

		publicURL, err := uploadToSupabaseWorkshop(mediaFile.Filename, file)
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload media to Supabase", err)
		}

		body.Media = publicURL
		mediaURL = publicURL
	}
	log.Printf("mediaURL: %v", mediaURL)

	if result := db.Create(&body); result.Error != nil {
		if err := deleteFileFromSupabase(body.Media); err != nil {
			log.Printf("Failed to delete media file after workshop creation failure: %v\n", err)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create workshop", result.Error)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Workshop created successfully", body)
}

func CreateProject(c *fiber.Ctx) error {
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
	fmt.Println("Retrieved userID:", userID)

	body := new(models.Project)
	body.UserID = userID
	if err := c.BodyParser(body); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}
	form, err := c.MultipartForm()
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Failed to parse form data", err)
	}
	files := form.File["media"]
	var mediaURL string
	if len(files) > 0 {
		mediaFile := files[0]

		file, err := mediaFile.Open()
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open media file", err)
		}
		defer file.Close()

		publicURL, err := uploadToSupabaseProjects(mediaFile.Filename, file)
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload media to Supabase", err)
		}

		body.Media = publicURL
		mediaURL = publicURL
	}
	log.Printf("mediaURL: %v", mediaURL)

	if result := db.Create(&body); result.Error != nil {
		if err := deleteFileFromSupabase(body.Media); err != nil {
			log.Printf("Failed to delete media file after project creation failure: %v\n", err)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create project", result.Error)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Project created successfully", body)
}

func uploadToSupabase(fileName string, fileContent io.Reader) (string, error) {
	bucketName := "file-buckets"
	folderName := "events"
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

	objectPath := fmt.Sprintf("%s/%s", folderName, fileName)
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

func deleteFileFromSupabase(filePath string) error {
	bucketName := "file-buckets"
	apiURL := os.Getenv("STORAGE_URL")
	authToken := "Bearer " + os.Getenv("SERVICE_ROLE_SECRET")

	if apiURL == "" {
		return fmt.Errorf("STORAGE_URL is not set in the environment variables")
	}

	requestURL := fmt.Sprintf("%s/object/%s/%s", apiURL, bucketName, filePath)

	req, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("Delete failed. Response Body: %s\n", string(respBody))
		return fmt.Errorf("delete failed with status: %s", resp.Status)
	}

	return nil
}

func uploadToSupabaseWorkshop(fileName string, fileContent io.Reader) (string, error) {
	bucketName := "file-buckets"
	folderName := "workshop"
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

	objectPath := fmt.Sprintf("%s/%s", folderName, fileName)
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

func uploadToSupabaseProjects(fileName string, fileContent io.Reader) (string, error) {
	bucketName := "file-buckets"
	folderName := "projects"
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

	objectPath := fmt.Sprintf("%s/%s", folderName, fileName)
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

func GetEventByID(c *fiber.Ctx) error {
	db := database.DB
	eventID := c.Params("id")

	var event models.Event
	if err := db.Table("events").Where("id = ?", eventID).First(&event).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.HandleError(c, fiber.StatusNotFound, "Event not found", nil)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Database query failed", err)
	}

	event.Media = getMediaURL(event.Media)

	return helpers.HandleSuccess(c, fiber.StatusOK, "Event details retrieved successfully", event)
}

func GetWorkshopByID(c *fiber.Ctx) error {
	db := database.DB
	workshopID := c.Params("id")

	var workshop models.Workshop
	if err := db.Table("workshops").Where("id = ?", workshopID).First(&workshop).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.HandleError(c, fiber.StatusNotFound, "Workshop not found", nil)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Database query failed", err)
	}

	workshop.Media = getMediaURL(workshop.Media)

	return helpers.HandleSuccess(c, fiber.StatusOK, "Workshop details retrieved successfully", workshop)
}

func GetProjectByID(c *fiber.Ctx) error {
	db := database.DB
	projectID := c.Params("id")

	var project models.Project
	if err := db.Table("projects").Where("id = ?", projectID).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.HandleError(c, fiber.StatusNotFound, "Project not found", nil)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Database query failed", err)
	}

	project.Media = getMediaURL(project.Media)

	return helpers.HandleSuccess(c, fiber.StatusOK, "Project details retrieved successfully", project)
}

func getMediaURL(filePath string) string {
	if filePath == "" {
		return ""
	}
	return filePath
}

func GetEventsFeed(c *fiber.Ctx) error {
	db := database.DB
	var events []models.Event
	err := db.Table("events").
		Order("date DESC").
		Limit(15).
		Find(&events).Error
	if err != nil {
		log.Println("Error fetching events feed:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch events feed",
		})
	}
	return c.JSON(fiber.Map{
		"data":    events,
		"status":  "success",
		"message": "Events feed retrieved successfully",
	})
}

func GetWorkshopsFeed(c *fiber.Ctx) error {
	db := database.DB
	var workshops []models.Workshop
	err := db.Table("workshops").
		Order("date DESC").
		Limit(15).
		Find(&workshops).Error
	if err != nil {
		log.Println("Error fetching workshops feed:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch workshops feed",
		})
	}
	return c.JSON(fiber.Map{
		"data":    workshops,
		"status":  "success",
		"message": "Workshops feed retrieved successfully",
	})
}

func GetProjectsFeed(c *fiber.Ctx) error {
	db := database.DB
	var projects []models.Project
	err := db.Table("projects").
		Order("start_date DESC").
		Limit(15).
		Find(&projects).Error
	if err != nil {
		log.Println("Error fetching projects feed:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch projects feed",
		})
	}
	return c.JSON(fiber.Map{
		"data":    projects,
		"status":  "success",
		"message": "Projects feed retrieved successfully",
	})
}
