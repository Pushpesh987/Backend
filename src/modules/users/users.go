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
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetProfile(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(string)
    fmt.Printf(" fetched user_id from jwt : %v",userID)
	

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

	skillQuery := `SELECT s.skill_name 
                   FROM user_skills us
                   JOIN skills s ON us.skill_id = s.skill_id
                   WHERE us.user_id = ?`

	var skills []string
	if err := db.Raw(skillQuery, userID).Scan(&skills).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch user skills", err)
	}
	profile.Skills = skills

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

// func CreateProfile(c *fiber.Ctx) error {
// 	db := database.DB
// 	userID := c.Locals("user_id").(string)

// 	body := new(models.User)
// 	if err := c.BodyParser(body); err != nil {
// 		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
// 	}

// 	getOrCreateID := func(tableName string, columnName string, value string) (string, error) {
// 		var record struct {
// 			ID string `gorm:"column:id"`
// 		}

// 		err := db.Table(tableName).Where(columnName+" = ?", value).First(&record).Error
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			newRecord := map[string]interface{}{columnName: value}
// 			result := db.Table(tableName).Create(&newRecord)
// 			if result.Error != nil {
// 				return "", result.Error
// 			}

// 			err = db.Table(tableName).Where(columnName+" = ?", value).Select("id").First(&record).Error
// 			if err != nil {
// 				return "", err
// 			}
// 		} else if err != nil {
// 			return "", err
// 		}

// 		return record.ID, nil
// 	}

// 	locationID, err := getOrCreateID("locations", "name", body.LocationName)
// 	if err != nil {
// 		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process location", err)
// 	}

// 	educationLevelID, err := getOrCreateID("education_levels", "level_name", body.EducationLevel)
// 	if err != nil {
// 		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process education level", err)
// 	}

// 	fieldOfStudyID, err := getOrCreateID("fields_of_study", "field_name", body.FieldOfStudy)
// 	if err != nil {
// 		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process field of study", err)
// 	}

// 	collegeNameID, err := getOrCreateID("colleges", "college_name", body.CollegeName)
// 	if err != nil {
// 		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process college name", err)
// 	}

// 	profileData := map[string]interface{}{
// 		"first_name":         body.FirstName,
// 		"last_name":          body.LastName,
// 		"age":                body.Age,
// 		"gender":             body.Gender,
// 		"dob":                body.Dob,
// 		"phone":              body.Phone,
// 		"email":              body.Email,
// 		"location_id":        locationID,
// 		"education_level_id": educationLevelID,
// 		"field_of_study_id":  fieldOfStudyID,
// 		"college_name_id":    collegeNameID,
// 	}

// 	if result := db.Table("users").Where("id = ?", userID).Updates(profileData); result.Error != nil {
// 		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update profile", result.Error)
// 	}

// 	return helpers.HandleSuccess(c, fiber.StatusOK, "User profile updated successfully", profileData)
// }

func UpdateProfile(c *fiber.Ctx) error {
    db := database.DB
    userID := c.Locals("user_id").(string)

    // Struct for storing profile data
    body := &models.User{}
    body.FirstName = c.FormValue("first_name")
    body.LastName = c.FormValue("last_name")
    body.Username = c.FormValue("username")
    body.Gender = c.FormValue("gender")
    body.Dob, _ = time.Parse("2006-01-02", c.FormValue("dob")) // Assuming date format as "YYYY-MM-DD"
    body.Phone = c.FormValue("phone")
    body.Email = c.FormValue("email")
    body.LocationName = c.FormValue("location_name")
    body.EducationLevel = c.FormValue("education_level")
    body.FieldOfStudy = c.FormValue("field_of_study")
    body.CollegeName = c.FormValue("college_name")

    updates := map[string]interface{}{}
    if body.FirstName != "" {
        updates["first_name"] = body.FirstName
    }
    if body.LastName != "" {
        updates["last_name"] = body.LastName
    }
    if body.Username != "" {
        updates["username"] = body.Username
    }
    if body.Gender != "" {
        updates["gender"] = body.Gender
    }
    if !body.Dob.IsZero() {
        updates["dob"] = body.Dob
    }
    if body.Phone != "" {
        updates["phone"] = body.Phone
    }
    if body.Email != "" {
        updates["email"] = body.Email
    }

    // Helper function for fetching or creating IDs
    getOrCreateID := func(tableName string, columnName string, value string) (string, error) {
        var record struct {
            ID string `gorm:"column:id"`
        }
        err := db.Table(tableName).Where(columnName+" = ?", value).First(&record).Error
        if errors.Is(err, gorm.ErrRecordNotFound) {
            newRecord := map[string]interface{}{columnName: value}
            result := db.Table(tableName).Create(&newRecord)
            if result.Error != nil {
                return "", result.Error
            }
            err = db.Table(tableName).Where(columnName+" = ?", value).Select("id").First(&record).Error
            if err != nil {
                return "", err
            }
        } else if err != nil {
            return "", err
        }
        return record.ID, nil
    }

    if body.LocationName != "" {
        locationID, err := getOrCreateID("locations", "name", body.LocationName)
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process location", err)
        }
        updates["location_id"] = locationID
    }

    if body.EducationLevel != "" {
        educationLevelID, err := getOrCreateID("education_levels", "level_name", body.EducationLevel)
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process education level", err)
        }
        updates["education_level_id"] = educationLevelID
    }

    if body.FieldOfStudy != "" {
        fieldOfStudyID, err := getOrCreateID("fields_of_study", "field_name", body.FieldOfStudy)
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process field of study", err)
        }
        updates["field_of_study_id"] = fieldOfStudyID
    }

    if body.CollegeName != "" {
        collegeNameID, err := getOrCreateID("colleges", "college_name", body.CollegeName)
        if err != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process college name", err)
        }
        updates["college_name_id"] = collegeNameID
    }
	file, err := c.FormFile("profile_photo")
	if err == nil {
		fileContent, err := file.Open()
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open file", err)
		}
		defer fileContent.Close()

		fileName := uuid.New().String() + "-" + file.Filename
		filePath := fmt.Sprintf("profile-photos/%s", fileName)
		publicURL, err := uploadToSupabase(filePath, fileContent)
		if err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload file to storage", err)
		}

		updates["profile_pic_url"] = publicURL
		updates["profile_pic_size"] = file.Size
		updates["profile_pic_storage_path"] = filePath
	}

    if len(updates) > 0 {
        if result := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates); result.Error != nil {
            return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update profile", result.Error)
        }
    }

    for _, skill := range c.FormValue("skills") {
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

    for _, interest := range c.FormValue("interests") {
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

    fmt.Println("Updates Map:", updates)
    fmt.Printf("Body after parsing: %+v\n", body)

    return helpers.HandleSuccess(c, fiber.StatusOK, "User profile updated successfully", nil)
}

func UploadProfilePhoto(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(string)

	file, err := c.FormFile("profile_photo")
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "File upload failed", err)
	}

	fileContent, err := file.Open()
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to open file", err)
	}
	defer fileContent.Close()
	fileName := uuid.New().String() + "-" + file.Filename
	filePath := fmt.Sprintf("profile-photos/%s", fileName)
	publicURL, err := uploadToSupabase(filePath, fileContent)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to upload file to storage", err)
	}

	updates := map[string]interface{}{
		"profile_pic_url":          publicURL,
		"profile_pic_size":         file.Size,
		"profile_pic_storage_path": filePath,
	}

	if result := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates); result.Error != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update profile photo metadata", result.Error)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Profile photo updated successfully", fiber.Map{"profile_photo_url": publicURL})
}

func uploadToSupabase(fileName string, fileContent io.Reader) (string, error) {
	bucketName := "file-buckets"
	folderName := "profile-photos"
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

func UpdateUserSkillsAndInterests(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(string)

	var input struct {
		Skills    []string `json:"skills"`
		Interests []string `json:"interests"`
	}
	if err := c.BodyParser(&input); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	processItems := func(items []string, table string, userTable string, idColumn string, nameColumn string) error {
		for _, item := range items {
			var id string

			query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", idColumn, table, nameColumn)
			if err := db.Raw(query, item).Scan(&id).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if id == "" {
				insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES ($1) RETURNING %s", table, nameColumn, idColumn)
				if err := db.Raw(insertQuery, item).Scan(&id).Error; err != nil {
					return err
				}
			}

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

	if err := processItems(input.Skills, "skills", "user_skills", "skill_id", "skill_name"); err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process skills", err)
	}

	if err := processItems(input.Interests, "interests", "user_interests", "interest_id", "interest_name"); err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to process interests", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Skills and interests updated successfully", nil)
}
