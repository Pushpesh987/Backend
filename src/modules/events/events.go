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
	"os"
	// "strings"
	// "errors"
	"net/http"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	// "gorm.io/gorm"
)

// CreateEvent handles the creation of an event.
func CreateEvent(c *fiber.Ctx) error {
    db := database.DB

    authID, ok := c.Locals("user_id").(string)
    if !ok || authID == "" {
        log.Println("Invalid or missing authID")
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing auth_id", nil)
    }

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

    userID, err := uuid.Parse(user.ID)
    if err != nil {
        log.Printf("Error parsing user ID as UUID: %v\n", err)
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
    }
    fmt.Println("Retrieved userID:", userID)

    // Prepare the event data
    body := new(models.Event)
    body.UserID = userID

    // Parse the event details
    if err := c.BodyParser(body); err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
    }

    // Get the media file if present
    form, err := c.MultipartForm()
    if err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Failed to parse form data", err)
    }
    files := form.File["media"]
    var mediaURL string
    if len(files) > 0 {
        mediaFile := files[0]

        // Open the file
        file, err := mediaFile.Open()
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open media file", err)
        }
        defer file.Close()

        // Upload to Supabase
        publicURL, err := uploadToSupabase(mediaFile.Filename, file)
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload media to Supabase", err)
        }

        // Set the media URL in the event
        body.Media = publicURL
        mediaURL = publicURL
    }
    log.Printf("mediaURL: %v",mediaURL)
    // Insert the event into the database
    if result := db.Create(&body); result.Error != nil {
		// Attempt to delete the uploaded file if event creation failed
		if err := deleteFileFromSupabase(body.Media); err != nil {
			log.Printf("Failed to delete media file after event creation failure: %v\n", err)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create event", result.Error)
	}
	

    return helpers.HandleSuccess(c, fiber.StatusCreated, "Event created successfully", body)
}

// CreateWorkshop handles the creation of a workshop.
func CreateWorkshop(c *fiber.Ctx) error {
    db := database.DB

    authID, ok := c.Locals("user_id").(string)
    if !ok || authID == "" {
        log.Println("Invalid or missing authID")
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing auth_id", nil)
    }

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

    userID, err := uuid.Parse(user.ID)
    if err != nil {
        log.Printf("Error parsing user ID as UUID: %v\n", err)
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
    }
    fmt.Println("Retrieved userID:", userID)

    // Prepare the workshop data
    body := new(models.Workshop)
    body.UserID = userID

    // Parse the workshop details from the request body
    if err := c.BodyParser(body); err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
    }

    // Convert the string duration to time.Duration if it exists
    if body.Duration != "" {
        // Convert Duration to a string for any future processing if needed
        parsedDuration := body.Duration
        // (Optional) You can log or process the parsed duration here if needed.
        log.Printf("Parsed duration: %v", parsedDuration)
    } else {
        // Handle the case where duration is not provided
        log.Println("Duration not provided")
    }

    // Get the media file if present
    form, err := c.MultipartForm()
    if err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Failed to parse form data", err)
    }
    files := form.File["media"]
    var mediaURL string
    if len(files) > 0 {
        mediaFile := files[0]

        // Open the file
        file, err := mediaFile.Open()
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open media file", err)
        }
        defer file.Close()

        // Upload to Supabase
        publicURL, err := uploadToSupabaseWorkshop(mediaFile.Filename, file)
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload media to Supabase", err)
        }

        // Set the media URL in the workshop
        body.Media = publicURL
        mediaURL = publicURL
    }
    log.Printf("mediaURL: %v", mediaURL)

    // Insert the workshop into the database
    if result := db.Create(&body); result.Error != nil {
        // Attempt to delete the uploaded file if workshop creation failed
        if err := deleteFileFromSupabase(body.Media); err != nil {
            log.Printf("Failed to delete media file after workshop creation failure: %v\n", err)
        }
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create workshop", result.Error)
    }

    return helpers.HandleSuccess(c, fiber.StatusCreated, "Workshop created successfully", body)
}

// CreateProject handles the creation of a project.
func CreateProject(c *fiber.Ctx) error {
    db := database.DB

    authID, ok := c.Locals("user_id").(string)
    if !ok || authID == "" {
        log.Println("Invalid or missing authID")
        return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing auth_id", nil)
    }

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

    userID, err := uuid.Parse(user.ID)
    if err != nil {
        log.Printf("Error parsing user ID as UUID: %v\n", err)
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
    }
    fmt.Println("Retrieved userID:", userID)

    // Prepare the project data
    body := new(models.Project)
    body.UserID = userID

    // Parse the project details from the request body
    if err := c.BodyParser(body); err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
    }

    // Get the media file if present
    form, err := c.MultipartForm()
    if err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Failed to parse form data", err)
    }
    files := form.File["media"]
    var mediaURL string
    if len(files) > 0 {
        mediaFile := files[0]

        // Open the file
        file, err := mediaFile.Open()
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open media file", err)
        }
        defer file.Close()

        // Upload to Supabase
        publicURL, err := uploadToSupabaseProjects(mediaFile.Filename, file)
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload media to Supabase", err)
        }

        // Set the media URL in the project
        body.Media = publicURL
        mediaURL = publicURL
    }
    log.Printf("mediaURL: %v", mediaURL)

    // Insert the project into the database
    if result := db.Create(&body); result.Error != nil {
        // Attempt to delete the uploaded file if project creation failed
        if err := deleteFileFromSupabase(body.Media); err != nil {
            log.Printf("Failed to delete media file after project creation failure: %v\n", err)
        }
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to create project", result.Error)
    }

    return helpers.HandleSuccess(c, fiber.StatusCreated, "Project created successfully", body)
}

func uploadToSupabase(fileName string, fileContent io.Reader) (string, error) {
	bucketName := "file-buckets"
	folderName := "events"             // Specify the folder name
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

	// Construct REST API URL for storage in the specified folder
	objectPath := fmt.Sprintf("%s/%s", folderName, fileName)
	requestURL := fmt.Sprintf("%s/object/%s/%s", apiURL, bucketName, objectPath)

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

	// Construct the public URL for the file in the folder
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

    // Construct the URL for deleting the file
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
	folderName := "workshop"             // Specify the folder name
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

	// Construct REST API URL for storage in the specified folder
	objectPath := fmt.Sprintf("%s/%s", folderName, fileName)
	requestURL := fmt.Sprintf("%s/object/%s/%s", apiURL, bucketName, objectPath)

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

	// Construct the public URL for the file in the folder
	publicURL := fmt.Sprintf("%s/object/public/%s/%s", apiURL, bucketName, objectPath)
	return publicURL, nil
}

func uploadToSupabaseProjects(fileName string, fileContent io.Reader) (string, error) {
	bucketName := "file-buckets"
	folderName := "projects"             // Specify the folder name
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

	// Construct REST API URL for storage in the specified folder
	objectPath := fmt.Sprintf("%s/%s", folderName, fileName)
	requestURL := fmt.Sprintf("%s/object/%s/%s", apiURL, bucketName, objectPath)

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

	// Construct the public URL for the file in the folder
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

    // Ensure the media URL is included if applicable
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

    // Include logic for getting media URL if present
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

    // Get the media URL if applicable
    project.Media = getMediaURL(project.Media)

    return helpers.HandleSuccess(c, fiber.StatusOK, "Project details retrieved successfully", project)
}

func getMediaURL(filePath string) string {
    if filePath == "" {
        return ""
    }
    // Construct the URL based on your storage bucket configuration

    return  filePath
}

// Get Events Feed
func GetEventsFeed(c *fiber.Ctx) error {
	db := database.DB
    var events []models.Event // Use the existing Event struct from your models package
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

// Get Workshops Feed
func GetWorkshopsFeed(c *fiber.Ctx) error {
	db := database.DB
    var workshops []models.Workshop // Use the existing Workshop struct from your models package
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

// Get Projects Feed
func GetProjectsFeed(c *fiber.Ctx) error {
	db := database.DB
    var projects []models.Project // Use the existing Project struct from your models package
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
