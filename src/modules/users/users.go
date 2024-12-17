package users

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetProfile(c *fiber.Ctx) error {
	db := database.DB
	userID := c.Locals("user_id").(string)
	fmt.Printf(" fetched user_id from jwt : %v", userID)

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
	folderName := "profile-photos" // The folder name where files should be uploaded
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

func capitalizeWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = strings.Title(word)
	}
	return strings.Join(words, " ")
}

func UpdateProfile(c *fiber.Ctx) error { 
	userID := c.Locals("user_id")
	if userID == nil {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "User ID not found in context", nil)
	}
	userIDStr, ok := userID.(string)
	if !ok {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid User ID", nil)
	}

	type UpdateProfileRequest struct {
		FirstName      string   `json:"first_name"`
		LastName       string   `json:"last_name"`
		Username       string   `json:username`
		Phone          string   `json:"phone"`
		Gender         string   `json:"gender"`
		Dob            time.Time   `json:"dob"`
		Age            int      `json:"age"`
		Skills         []string `json:"skills"`
		Interests      []string `json:"interests"`
		Location       string   `json:"location_name"`
		EducationLevel string   `json:"education_level"`
		FieldOfStudy   string   `json:"field_of_study"`
		CollegeName    string   `json:"college_name"`
	}

	var request UpdateProfileRequest

	if err := c.BodyParser(&request); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid request data", err)
	}

	db := database.DB
	var user models.User
	err := db.Raw("SELECT * FROM users WHERE id = $1", userIDStr).Scan(&user).Error
	if err != nil {
		if err.Error() == "record not found" {
			return helpers.HandleError(c, fiber.StatusNotFound, "User not found", err)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Database error", err)
	}

if request.FirstName != "" {
    log.Printf("Updating FirstName: Old Value = %s, New Value = %s", user.FirstName, request.FirstName)
    user.FirstName = request.FirstName
    if err := db.Model(&user).Update("first_name", user.FirstName).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's first name", err)
    }
}

if request.LastName != "" {
    log.Printf("Updating LastName: Old Value = %s, New Value = %s", user.LastName, request.LastName)
    user.LastName = request.LastName
    if err := db.Model(&user).Update("last_name", user.LastName).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's last name", err)
    }
}

if !request.Dob.IsZero(){
    log.Printf("Updating DOB: Old Value = %s, New Value = %s", user.Dob, request.Dob)
    user.Dob = request.Dob
    if err := db.Model(&user).Update("dob", user.Dob).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's date of birth", err)
    }
}

if request.Age != 0 {
    log.Printf("Updating Age: Old Value = %d, New Value = %d", user.Age, request.Age)
    user.Age = request.Age
    if err := db.Model(&user).Update("age", user.Age).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's age", err)
    }
}

if request.Phone != "" {
    log.Printf("Updating Phone: Old Value = %s, New Value = %s", user.Phone, request.Phone)
    user.Phone = request.Phone
    if err := db.Model(&user).Update("phone", user.Phone).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's phone", err)
    }
}

if request.Gender != "" {
    log.Printf("Updating Gender: Old Value = %s, New Value = %s", user.Gender, request.Gender)
    user.Gender = request.Gender
    if err := db.Model(&user).Update("gender", user.Gender).Error; err != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's gender", err)
    }
}

	if request.Location != "" {
		request.Location = capitalizeWords(request.Location)

		var location models.Location

		err := db.Where("name = ?", request.Location).First(&location).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				location = models.Location{
					ID:           uuid.New(),
					LocationName: request.Location,
				}

				if err := db.Create(&location).Error; err != nil {
					return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to insert new location", err)
				}
			} else {
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to retrieve location information", err)
			}
		}

		if err := db.Model(&user).Update("location_id", location.ID).Error; err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's location information", err)
		}
		log.Printf("Location updated: Old LocationID = %v, New LocationID = %v", user.LocationID, location.ID)
	}

	if request.EducationLevel != "" {
		request.EducationLevel = capitalizeWords(request.EducationLevel)

		var educationLevel models.EducationLevel

		err := db.Where("level_name = ?", request.EducationLevel).First(&educationLevel).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				educationLevel = models.EducationLevel{
					ID:   uuid.New(),
					Name: request.EducationLevel,
				}

				if err := db.Create(&educationLevel).Error; err != nil {
					return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to insert new education level", err)
				}
			} else {
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to retrieve education level information", err)
			}
		}
		log.Printf("EducationLevel updated: Old ID = %v, New ID = %v", user.EducationLevelID, educationLevel.ID)

		if err := db.Model(&user).Update("education_level_id", educationLevel.ID).Error; err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's education level information", err)
		}
	}

	if request.FieldOfStudy != "" {
		request.FieldOfStudy = capitalizeWords(request.FieldOfStudy)

		var fieldOfStudy models.FieldOfStudy

		err := db.Where("field_name = ?", request.FieldOfStudy).First(&fieldOfStudy).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				fieldOfStudy = models.FieldOfStudy{
					ID:   uuid.New(),
					Name: request.FieldOfStudy,
				}
				if err := db.Create(&fieldOfStudy).Error; err != nil {
					return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to insert new field of study", err)
				}
			} else {
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to retrieve field of study information", err)
			}
		}
		log.Printf("FieldOfStudy updated: Old ID = %v, New ID = %v", user.FieldOfStudyID, fieldOfStudy.ID)

		if err := db.Model(&user).Update("field_of_study_id", fieldOfStudy.ID).Error; err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's field of study information", err)
		}
	}

	if request.CollegeName != "" {
		request.CollegeName = capitalizeWords(request.CollegeName)

		var college models.College

		err := db.Where("college_name = ?", request.CollegeName).First(&college).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				college = models.College{
					ID:          uuid.New(),
					CollegeName: request.CollegeName,
				}
				if err := db.Create(&college).Error; err != nil {
					return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to insert new college", err)
				}
			} else {
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to retrieve college information", err)
			}
		}

		if err := db.Model(&user).Update("college_name_id", college.ID).Error; err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to update user's college information", err)
		}
		if err := db.Save(&user).Error; err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to save user after update", err)
		}
		log.Printf("CollegeName updated: Old ID = %v, New ID = %v", user.CollegeNameID, college.ID)
		log.Printf("Updated User: %+v", user)
	}

	if len(request.Skills) > 0 {
		requestedSkills := make(map[string]bool)
		for _, skillName := range request.Skills {
			skillName = capitalizeWords(skillName)
			requestedSkills[skillName] = true
		}

		var currentSkills []models.Skill
		if err := db.Joins("JOIN user_skills us ON us.skill_id = skills.skill_id").
			Where("us.user_id = ?", user.ID).
			Find(&currentSkills).Error; err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch current skills", err)
		}

		currentSkillMap := make(map[string]uuid.UUID)
		for _, skill := range currentSkills {
			currentSkillMap[skill.SkillName] = skill.SkillID
		}

		skillsToAdd := []string{}
		skillsToRemove := []uuid.UUID{}

		for skillName := range requestedSkills {
			if _, exists := currentSkillMap[skillName]; !exists {
				skillsToAdd = append(skillsToAdd, skillName)
			}
		}

		for skillName, skillID := range currentSkillMap {
			if _, exists := requestedSkills[skillName]; !exists {
				skillsToRemove = append(skillsToRemove, skillID)
			}
		}

		for _, skillName := range skillsToAdd {
			var skill models.Skill
			err := db.Where("skill_name = ?", skillName).First(&skill).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					skill = models.Skill{
						SkillID:   uuid.New(),
						SkillName: skillName,
					}
					if err := db.Create(&skill).Error; err != nil {
						return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to add skill", err)
					}
				} else {
					return helpers.HandleError(c, fiber.StatusInternalServerError, "Database error while searching for skill", err)
				}
			}

			if err := db.Exec("INSERT INTO user_skills (user_id, skill_id) VALUES (?, ?) ON CONFLICT DO NOTHING", user.ID, skill.SkillID).Error; err != nil {
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to map skill to user", err)
			}
		}

		if len(skillsToRemove) > 0 {
			if err := db.Exec("DELETE FROM user_skills WHERE user_id = ? AND skill_id IN ?", user.ID, skillsToRemove).Error; err != nil {
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to remove skills from user", err)
			}
		}
		log.Printf("Skills updated: Adding = %v, Removing = %v", skillsToAdd, skillsToRemove)
	}

	if len(request.Interests) > 0 {
		requestedInterests := make(map[string]bool)
		for _, interestName := range request.Interests {
			interestName = capitalizeWords(interestName)
			requestedInterests[interestName] = true
		}

		var currentInterests []models.Interest
		if err := db.Joins("JOIN user_interests ui ON ui.interest_id = interests.interest_id").
			Where("ui.user_id = ?", user.ID).
			Find(&currentInterests).Error; err != nil {
			return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch current interests", err)
		}

		currentInterestMap := make(map[string]uuid.UUID)
		for _, interest := range currentInterests {
			currentInterestMap[interest.InterestName] = interest.InterestID
		}

		interestsToAdd := []string{}
		interestsToRemove := []uuid.UUID{}

		for interestName := range requestedInterests {
			if _, exists := currentInterestMap[interestName]; !exists {
				interestsToAdd = append(interestsToAdd, interestName)
			}
		}

		for interestName, interestID := range currentInterestMap {
			if _, exists := requestedInterests[interestName]; !exists {
				interestsToRemove = append(interestsToRemove, interestID)
			}
		}

		for _, interestName := range interestsToAdd {
			var interest models.Interest
			err := db.Where("interest_name = ?", interestName).First(&interest).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					interest = models.Interest{
						InterestID:   uuid.New(),
						InterestName: interestName,
					}
					if err := db.Create(&interest).Error; err != nil {
						return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to add interest", err)
					}
				} else {
					return helpers.HandleError(c, fiber.StatusInternalServerError, "Database error while searching for interest", err)
				}
			}

			if err := db.Exec("INSERT INTO user_interests (user_id, interest_id) VALUES (?, ?) ON CONFLICT DO NOTHING", user.ID, interest.InterestID).Error; err != nil {
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to map interest to user", err)
			}
		}

		if len(interestsToRemove) > 0 {
			if err := db.Exec("DELETE FROM user_interests WHERE user_id = ? AND interest_id IN ?", user.ID, interestsToRemove).Error; err != nil {
				return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to remove interests from user", err)
			}
		}
		log.Printf("Interests updated: Adding = %v, Removing = %v", interestsToAdd, interestsToRemove)
	}
	return helpers.HandleSuccess(c, fiber.StatusOK, "User profile updated successfully", user)
}

func GetAllLocationNames(c *fiber.Ctx) error {
	db := database.DB

	locationQuery := `SELECT l.id AS id, l.name AS location_name FROM locations l`

	type Location struct {
		ID           uuid.UUID `json:"id"`
		LocationName string    `json:"location_name"`
	}

	var locations []Location

	if err := db.Raw(locationQuery).Scan(&locations).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch all locations", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "All locations retrieved successfully", locations)
}

func GetAllSkills(c *fiber.Ctx) error {
	db := database.DB

	type Skill struct {
		SkillName string `json:"skill_name"`
	}

	var skills []Skill

	skillQuery := `SELECT skill_name FROM skills`

	if err := db.Raw(skillQuery).Scan(&skills).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch all skills", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "All skills retrieved successfully", skills)
}

func GetAllInterests(c *fiber.Ctx) error {
	db := database.DB

	type Interest struct {
		InterestName string `json:"interest_name"`
	}

	var interests []Interest

	interestQuery := `SELECT interest_name FROM interests`

	if err := db.Raw(interestQuery).Scan(&interests).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch all interests", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "All interests retrieved successfully", interests)
}

func GetAllFieldsOfStudy(c *fiber.Ctx) error {
	db := database.DB

	type FieldOfStudy struct {
		FieldName string `json:"field_name"`
	}

	var fieldsOfStudy []FieldOfStudy

	fieldQuery := `SELECT field_name FROM fields_of_study`

	if err := db.Raw(fieldQuery).Scan(&fieldsOfStudy).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch all fields of study", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "All fields of study retrieved successfully", fieldsOfStudy)
}

func GetAllEducationLevels(c *fiber.Ctx) error {
	db := database.DB

	type EducationLevel struct {
		LevelName string `json:"level_name"`
	}

	var educationLevels []EducationLevel

	educationQuery := `SELECT level_name FROM education_levels`

	if err := db.Raw(educationQuery).Scan(&educationLevels).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch all education levels", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "All education levels retrieved successfully", educationLevels)
}

func GetAllColleges(c *fiber.Ctx) error {
	db := database.DB

	type College struct {
		CollegeName string `json:"college_name"`
	}

	var colleges []College

	collegeQuery := `SELECT college_name FROM colleges`

	if err := db.Raw(collegeQuery).Scan(&colleges).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch all colleges", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "All colleges retrieved successfully", colleges)
}

func SearchUsers(c *fiber.Ctx) error {
    db := database.DB

    searchQuery := c.Query("query", "")
    if searchQuery == "" {
        return helpers.HandleError(c, fiber.StatusBadRequest, "Search query is required", nil)
    }

    searchTerm := "%" + searchQuery + "%"

	var results []struct {
        UserID   uuid.UUID `json:"id" gorm:"column:id"`  
        Username string    `json:"username" gorm:"column:username"` 
    }

    result := db.Table("users").Select("id, username").
        Where("LOWER(first_name) LIKE ? OR LOWER(username) LIKE ? OR LOWER(email) LIKE ?", 
        searchTerm, searchTerm, searchTerm).Find(&results)

    if result.Error != nil {
        return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch search results", result.Error)
    }

    if len(results) == 0 {
        return helpers.HandleError(c, fiber.StatusNotFound, "No matching users found", nil)
    }

    return helpers.HandleSuccess(c, fiber.StatusOK, "Search completed successfully", results)
}

func GetProfileByID(c *fiber.Ctx) error {
    db := database.DB
    userID := c.Params("id") 

    if userID == "" {
        return helpers.HandleError(c, fiber.StatusBadRequest, "User ID is required", nil)
    }

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

