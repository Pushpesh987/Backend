package questions

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetDailyQuestions(c *fiber.Ctx) error {
	db := database.DB
	var questions []models.Question

	err := db.Where("question_type = ?", "Daily").
		Order("created_at DESC").
		Limit(5).
		Find(&questions).Error

	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch daily questions", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Daily questions fetched successfully", questions)
}

func GetSkillQuestions(c *fiber.Ctx) error {
	db := database.DB
	var questions []models.Question

	err := db.Where("question_type = ?", "Skill").
		Order("created_at DESC").
		Limit(3).
		Find(&questions).Error

	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch skill questions", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Skill questions fetched successfully", questions)
}

func GetBonusQuestions(c *fiber.Ctx) error {
	db := database.DB
	var questions []models.Question

	err := db.Where("question_type = ?", "Bonus").
		Order("created_at DESC").
		Limit(2).
		Find(&questions).Error

	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch bonus questions", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Bonus questions fetched successfully", questions)
}

func SubmitAnswer(c *fiber.Ctx) error {
	db := database.DB

	var input struct {
		QuestionID int  `json:"question_id"`
		IsCorrect  bool `json:"is_correct"`
	}

	if err := c.BodyParser(&input); err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid input data", err)
	}

	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
	}

	userID, err := uuid.Parse(userId)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	quizAttempt := models.QuizAttempt{
		UserID:     userID,
		QuestionID: input.QuestionID,
		IsCorrect:  input.IsCorrect,
	}

	if err := db.Create(&quizAttempt).Error; err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to store quiz attempt", err)
	}

	return helpers.HandleSuccess(c, fiber.StatusCreated, "Answer submitted successfully", quizAttempt)
}



