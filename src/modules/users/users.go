package users

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

// GetProfile retrieves the profile of the authenticated user.
func GetProfile(c *fiber.Ctx) error {
    db := database.DB
    authID := c.Locals("user_id").(string) // This is actually the auth_id from JWT

    // Fetch the user_id using auth_id
    var userID string
    if err := db.Raw("SELECT id FROM users WHERE auth_id = ?", authID).Scan(&userID).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusNotFound, "User not found", err)
    }

    // Query for basic profile details
    profileQuery := `SELECT u.*, 
                            COALESCE(l.name, '') AS location_name, 
                            COALESCE(e.level_name, '') AS education_level, 
                            COALESCE(f.field_name, '') AS field_of_study, 
                            COALESCE(c.college_name, '') AS college_name 
                     FROM users u
                     LEFT JOIN locations l ON u.location_id = l.id
                     LEFT JOIN education_levels e ON u.education_level_id = e.id
                     LEFT JOIN fields_of_study f ON u.field_of_study_id = f.id
                     LEFT JOIN colleges c ON u.college_name_id = c.id
                     WHERE u.id = ?`

    profile := struct {
        models.User
        LocationName   string   `json:"location_name"`
        EducationLevel string   `json:"education_level"`
        FieldOfStudy   string   `json:"field_of_study"`
        CollegeName    string   `json:"college_name"`
        Skills         []string `json:"skills"`
        Interests      []string `json:"interests"`
    }{}

    if err := db.Raw(profileQuery, userID).Scan(&profile).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusNotFound, "User profile not found", err)
    }

    // Query for user skills
    skillQuery := `SELECT s.skill_name 
                   FROM user_skills us
                   JOIN skills s ON us.skill_id = s.skill_id
                   WHERE us.user_id = ?`

    var skills []string
    if err := db.Raw(skillQuery, userID).Scan(&skills).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch user skills", err)
    }
    profile.Skills = skills

    // Query for user interests
    interestQuery := `SELECT i.interest_name 
                      FROM user_interests ui
                      JOIN interests i ON ui.interest_id = i.interest_id
                      WHERE ui.user_id = ?`

    var interests []string
    if err := db.Raw(interestQuery, userID).Scan(&interests).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch user interests", err)
    }
    profile.Interests = interests

    return helpers.HandleSuccess(c, fiber.StatusOK, "User profile retrieved successfully", profile)
}

// CreateProfile handles the creation of the user's profile.
func CreateProfile(c *fiber.Ctx) error {
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

    body := new(models.User)
    if err := c.BodyParser(body); err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
    }

    // Helper function to get or create ID for a table
    getOrCreateID := func(tableName string, columnName string, value string) (string, error) {
        var record struct {
            ID string `gorm:"column:id"`
        }

        // Try to find the record
        err := db.Table(tableName).Where(columnName+" = ?", value).First(&record).Error
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // Record not found, create a new one
            newRecord := map[string]interface{}{columnName: value}
            result := db.Table(tableName).Create(&newRecord)
            if result.Error != nil {
                return "", result.Error
            }

            // Fetch the ID of the newly created record
            err = db.Table(tableName).Where(columnName+" = ?", value).Select("id").First(&record).Error
            if err != nil {
                return "", err
            }
        } else if err != nil {
            return "", err
        }

        return record.ID, nil
    }

    locationID, err := getOrCreateID("locations", "name", body.LocationName)
    if err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process location", err)
    }

    educationLevelID, err := getOrCreateID("education_levels", "level_name", body.EducationLevelName)
    if err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process education level", err)
    }

    fieldOfStudyID, err := getOrCreateID("fields_of_study", "field_name", body.FieldOfStudyName)
    if err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process field of study", err)
    }

    collegeNameID, err := getOrCreateID("colleges", "college_name", body.CollegeName)
    if err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process college name", err)
    }

    // Update the user's profile
    profileData := map[string]interface{}{
        "first_name":         body.FirstName,
        "last_name":          body.LastName,
        "age":                body.Age,
        "gender":             body.Gender,
        "dob":                body.Dob,
        "phone":              body.Phone,
        "email":              body.Email,
        "location_id":        locationID,
        "education_level_id": educationLevelID,
        "field_of_study_id":  fieldOfStudyID,
        "college_name_id":    collegeNameID,
    }

    if result := db.Table("users").Where("id = ?", user.ID).Updates(profileData); result.Error != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update profile", result.Error)
    }

    return helpers.HandleSuccess(c, fiber.StatusOK, "User profile updated successfully", profileData)
}

// UpdateProfile updates the profile of the authenticated user.

func UpdateProfile(c *fiber.Ctx) error {
    db := database.DB
    userID := c.Locals("user_id").(string)

    // Parse the request body
    body := struct {
        models.User
        Skills    []string  `json:"skills"`
        Interests []string  `json:"interests"`
        LocationID        uuid.UUID `json:"location_id"`
        EducationLevelID  uuid.UUID `json:"education_level_id"`
        FieldOfStudyID    uuid.UUID `json:"field_of_study_id"`
        CollegeNameID     uuid.UUID `json:"college_name_id"`
    }{}

    if err := c.BodyParser(&body); err != nil {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
    }

    // Update basic user details
    updates := map[string]interface{}{
        "first_name":         body.FirstName,
        "last_name":          body.LastName,
        "gender":             body.Gender,
        "dob":                body.Dob,
        "phone":              body.Phone,
        "email":              body.Email,
        "location_id":        body.LocationID,
        "education_level_id": body.EducationLevelID,
        "field_of_study_id":  body.FieldOfStudyID,
        "college_name_id":    body.CollegeNameID,
    }

    if result := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates); result.Error != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update profile", result.Error)
    }

    // Handle Location
    if body.LocationID != uuid.Nil {
        var locationID uuid.UUID
        locationCheckQuery := `INSERT INTO locations (name) 
                               VALUES (?) 
                               ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id`
        if err := db.Raw(locationCheckQuery, body.LocationID).Scan(&locationID).Error; err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process location", err)
        }

        // Update the profile with the new location ID
        updates["location_id"] = locationID
        if result := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates); result.Error != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update location", result.Error)
        }
    }

    // Handle Education Level
    if body.EducationLevelID != uuid.Nil {
        var educationLevelID uuid.UUID
        educationLevelCheckQuery := `INSERT INTO education_levels (level_name) 
                                     VALUES (?) 
                                     ON CONFLICT (level_name) DO UPDATE SET level_name=EXCLUDED.level_name RETURNING id`
        if err := db.Raw(educationLevelCheckQuery, body.EducationLevelID).Scan(&educationLevelID).Error; err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process education level", err)
        }

        // Update the profile with the new education level ID
        updates["education_level_id"] = educationLevelID
        if result := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates); result.Error != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update education level", result.Error)
        }
    }

    // Handle Field of Study
    if body.FieldOfStudyID != uuid.Nil {
        var fieldOfStudyID uuid.UUID
        fieldOfStudyCheckQuery := `INSERT INTO fields_of_study (field_name) 
                                   VALUES (?) 
                                   ON CONFLICT (field_name) DO UPDATE SET field_name=EXCLUDED.field_name RETURNING id`
        if err := db.Raw(fieldOfStudyCheckQuery, body.FieldOfStudyID).Scan(&fieldOfStudyID).Error; err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process field of study", err)
        }

        // Update the profile with the new field of study ID
        updates["field_of_study_id"] = fieldOfStudyID
        if result := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates); result.Error != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update field of study", result.Error)
        }
    }

    // Handle skills
    for _, skill := range body.Skills {
        var skillID string
        skillCheckQuery := `INSERT INTO skills (skill_name) 
                            VALUES (?) 
                            ON CONFLICT (skill_name) DO UPDATE SET skill_name=EXCLUDED.skill_name RETURNING skill_id`
        if err := db.Raw(skillCheckQuery, skill).Scan(&skillID).Error; err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process skill", err)
        }

        userSkillQuery := `INSERT INTO user_skills (user_id, skill_id) VALUES (?, ?) 
                           ON CONFLICT (user_id, skill_id) DO NOTHING`
        if err := db.Exec(userSkillQuery, userID, skillID).Error; err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to link skill", err)
        }
    }

    // Handle interests
    for _, interest := range body.Interests {
        var interestID string
        interestCheckQuery := `INSERT INTO interests (interest_name) 
                               VALUES (?) 
                               ON CONFLICT (interest_name) DO UPDATE SET interest_name=EXCLUDED.interest_name RETURNING interest_id`
        if err := db.Raw(interestCheckQuery, interest).Scan(&interestID).Error; err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process interest", err)
        }

        userInterestQuery := `INSERT INTO user_interests (user_id, interest_id) VALUES (?, ?) 
                              ON CONFLICT (user_id, interest_id) DO NOTHING`
        if err := db.Exec(userInterestQuery, userID, interestID).Error; err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to link interest", err)
        }
    }

    return helpers.HandleSuccess(c, fiber.StatusOK, "User profile updated successfully", nil)
}


// UploadProfilePhoto handles the upload of a profile photo and updates the user record.
func UploadProfilePhoto(c *fiber.Ctx) error {
    db := database.DB
    userID := c.Locals("user_id").(string)  // Extracting user_id from JWT

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

    // Update user's profile photo URL and metadata in DB
    updates := map[string]interface{}{
        "profile_pic_url":           publicURL,
        "profile_pic_size":          file.Size,
        "profile_pic_content_type":  file.Header.Get("Content-Type"),
        "profile_pic_storage_path":  filePath,
    }

    // Use auth_id for querying user and updating the profile photo
    if result := db.Model(&models.User{}).Where("auth_id = ?", userID).Updates(updates); result.Error != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update profile photo metadata", result.Error)
    }

    return helpers.HandleSuccess(c, fiber.StatusOK, "Profile photo updated successfully", fiber.Map{"profile_photo_url": publicURL})
}

// uploadToSupabase uploads a file to Supabase storage and returns its public URL.
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

func UpdateUserSkillsAndInterests(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(string)

	// Parse input
	var input struct {
		Skills    []string `json:"skills"`
		Interests []string `json:"interests"`
	}
	if err := c.BodyParser(&input); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	// Helper to process skills or interests
	processItems := func(items []string, table string, userTable string, idColumn string, nameColumn string) error {
		for _, item := range items {
			var id string

			// Check if the item exists
			query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", idColumn, table, nameColumn)
			if err := db.Raw(query, item).Scan(&id).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			// Insert into the table if it doesn't exist
			if id == "" {
				insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES ($1) RETURNING %s", table, nameColumn, idColumn)
				if err := db.Raw(insertQuery, item).Scan(&id).Error; err != nil {
					return err
				}
			}

			// Insert into user-specific table
			userInsertQuery := fmt.Sprintf(
				"INSERT INTO %s (user_id, %s) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				userTable, idColumn,
			)
			if err := db.Exec(userInsertQuery, userID, id).Error; err != nil {
				return err
			}
		}
		return nil
	}

	// Process skills
	if err := processItems(input.Skills, "skills", "user_skills", "skill_id", "skill_name"); err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process skills", err)
	}

	// Process interests
	if err := processItems(input.Interests, "interests", "user_interests", "interest_id", "interest_name"); err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process interests", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Skills and interests updated successfully", nil)
}
