package events

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const dateFormat = "2006-01-02"

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

	body := new(models.Event)
	body.UserID = userID

	form, err := c.MultipartForm()
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Failed to parse form data", err)
	}

	body.Title = form.Value["title"][0]
	body.Theme = form.Value["theme"][0]
	body.Description = form.Value["description"][0]
	body.Location = form.Value["location"][0]
	body.OrganizerName = form.Value["organizer_name"][0]
	body.OrganizerContact = form.Value["organizer_contact"][0]
	body.Status = form.Value["status"][0]

	// Debugging: Log received date and registration_deadline
	log.Printf("Received date: %s, registration_deadline: %s\n", form.Value["date"][0], form.Value["registration_deadline"][0])

	// Handle attendee count
	if count, err := strconv.Atoi(form.Value["attendee_count"][0]); err == nil {
		body.AttendeeCount = count
	}

	// Handle entry fee and prize pool (same as before)
	if entryFee, err := strconv.Atoi(form.Value["entry_fee"][0]); err == nil {
		body.EntryFee = entryFee
	} else {
		log.Printf("Error converting entry_fee: %v\n", err)
	}

	if prizePool, err := strconv.Atoi(form.Value["prize_pool"][0]); err == nil {
		body.PrizePool = prizePool
	} else {
		log.Printf("Error converting prize_pool: %v\n", err)
	}

	// Handle tags (optional field)
	if len(form.Value["tags"]) > 0 {
		body.Tags = form.Value["tags"][0]
	}

	// Parse date (Handle date without time)
	if dateStr, ok := form.Value["date"]; ok && len(dateStr) > 0 {
		parsedDate, err := time.Parse("2006-01-02", dateStr[0]) // Updated format for date only
		if err == nil {
			body.Date = parsedDate
		} else {
			log.Printf("Error converting date: %v\n", err)
		}
	}

	// Parse registration_deadline (Ensure it is in valid format before parsing)
	if regDeadlineStr, ok := form.Value["registration_deadline"]; ok && len(regDeadlineStr) > 0 {
		// Fix any common mistakes like "1o" -> "10"
		fixedDeadline := strings.Replace(regDeadlineStr[0], "o", "0", -1)

		parsedDeadline, err := time.Parse("2006-01-02", fixedDeadline) // Use the same format for date without time
		if err == nil {
			body.RegistrationDeadline = parsedDeadline
		} else {
			log.Printf("Error converting registration_deadline: %v\n", err)
		}
	}

	// Handle media upload
	files := form.File["media"]
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
	}

	// Insert into database
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

	// Extract and validate user ID from context
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

	// Parse form-data
	form, err := c.MultipartForm()
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Failed to parse form data", err)
	}

	// Map form-data fields to the Workshop struct
	body := new(models.Workshop)
	body.UserID = userID

	// Required fields
	body.Title = form.Value["title"][0]
	body.Date, _ = time.Parse(time.RFC3339, form.Value["date"][0]) // Parse date (format: YYYY-MM-DDTHH:MM:SSZ)
	body.Status = form.Value["status"][0]

	// Optional fields
	if len(form.Value["description"]) > 0 {
		body.Description = form.Value["description"][0]
	}
	if len(form.Value["duration"]) > 0 {
		body.Duration = form.Value["duration"][0]
	}
	if len(form.Value["location"]) > 0 {
		body.Location = form.Value["location"][0]
	}
	if len(form.Value["entry_fee"]) > 0 {
		body.EntryFee = form.Value["entry_fee"][0]
	}
	if len(form.Value["instructor_info"]) > 0 {
		body.InstructorInfo = form.Value["instructor_info"][0]
	}
	if len(form.Value["participant_limit"]) > 0 {
		// Convert the string to int
		participantLimitStr := form.Value["participant_limit"][0]
		participantLimit, err := strconv.Atoi(participantLimitStr)
		if err != nil {
			log.Println("Error converting participant_limit to int:", err)
			return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid participant_limit", err)
		}

		// Assign the converted integer value to the struct field
		body.ParticipantLimit = participantLimit
	}

	if len(form.Value["tags"]) > 0 {
		body.Tags = form.Value["tags"][0]
	}
	if len(form.Value["registration_link"]) > 0 {
		body.RegistrationLink = form.Value["registration_link"][0]
	}

	// Handle media file upload (if provided)
	files := form.File["media"]
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
		log.Printf("Media uploaded successfully: %s", publicURL)
	}

	// Save to database
	if result := db.Create(&body); result.Error != nil {
		if body.Media != "" {
			if err := deleteFileFromSupabase(body.Media); err != nil {
				log.Printf("Failed to delete media file after workshop creation failure: %v\n", err)
			}
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

	body := new(models.Project)
	body.UserID = userID

	if err := c.BodyParser(body); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Failed to parse form data", err)
	}

	log.Printf("Form values: %v", form.Value)

	if len(form.Value["team_members"]) > 0 {
		body.TeamMembers = form.Value["team_members"][0]
	}
	if len(form.Value["project_link"]) > 0 {
		body.ProjectLink = form.Value["project_link"][0]
	}

	// Handle start_date
	if len(form.Value["start_date"]) > 0 {
		if startDate, err := time.Parse(dateFormat, form.Value["start_date"][0]); err == nil {
			body.StartDate = startDate
		} else {
			log.Printf("Error parsing start_date: %v\n", err)
		}
	}

	// Handle end_date
	if len(form.Value["end_date"]) > 0 {
		if endDate, err := time.Parse(dateFormat, form.Value["end_date"][0]); err == nil {
			body.EndDate = endDate
		} else {
			log.Printf("Error parsing end_date: %v\n", err)
		}
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

	// Create the project in the database
	if result := db.Create(&body); result.Error != nil {
		// Clean up Supabase file if project creation fails
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

	// Generate a unique file name using timestamp
	timestamp := time.Now().UnixNano() // nanosecond precision
	ext := filepath.Ext(fileName)      // Extract file extension
	baseName := strings.TrimSuffix(fileName, ext)
	uniqueFileName := fmt.Sprintf("%s_%d%s", baseName, timestamp, ext)

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

func uploadToSupabaseProjects(fileName string, fileContent io.Reader) (string, error) {
	bucketName := "file-buckets"
	folderName := "projects"
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
