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
	var todaysQuestions []models.Question
	var answeredQuestionIDs []int
	var remainingQuestions []models.Question

	err := db.Table("questions").
		Select("question_id, question_text, options, correct_answer, difficulty, points, multiplier, question_type, created_at").
		Where("question_type = ?", "Daily").
		Order("created_at DESC").
		Limit(5).
		Find(&todaysQuestions).Error

	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch today's questions", err)
	}

	var todaysQuestionIDs []int
	for _, question := range todaysQuestions {
		todaysQuestionIDs = append(todaysQuestionIDs, question.QuestionID)
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
	}

	err = db.Table("quiz_attempts").
		Select("question_id").
		Where("user_id = ?", userID).
		Where("question_id IN ?", todaysQuestionIDs). 
		Find(&answeredQuestionIDs).Error

	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch answered questions", err)
	}

	answeredQuestionsMap := make(map[int]bool)
	for _, id := range answeredQuestionIDs {
		answeredQuestionsMap[id] = true
	}

	for _, question := range todaysQuestions {
		if !answeredQuestionsMap[question.QuestionID] {
			remainingQuestions = append(remainingQuestions, question)
		}
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Daily questions fetched successfully", remainingQuestions)
}

func GetSkillQuestions(c *fiber.Ctx) error {
	db := database.DB
	var questions []models.Question
	var answeredQuestionIDs []int
	var remainingQuestions []models.Question

	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
	}

	userID, err := uuid.Parse(userId)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	err = db.Table("questions").
		Select("question_id, question_text, options, correct_answer, difficulty, points, multiplier, question_type, created_at").
		Where("question_type = ?", "Skill").
		Order("created_at DESC").
		Limit(3).
		Find(&questions).Error

	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch skill questions", err)
	}

	var skillQuestionIDs []int
	for _, question := range questions {
		skillQuestionIDs = append(skillQuestionIDs, question.QuestionID)
	}

	err = db.Table("quiz_attempts").
		Select("question_id").
		Where("user_id = ?", userID).
		Where("question_id IN ?", skillQuestionIDs). 
		Find(&answeredQuestionIDs).Error

	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch answered skill questions", err)
	}

	answeredQuestionsMap := make(map[int]bool)
	for _, id := range answeredQuestionIDs {
		answeredQuestionsMap[id] = true
	}

	for _, question := range questions {
		if !answeredQuestionsMap[question.QuestionID] {
			remainingQuestions = append(remainingQuestions, question)
		}
	}

	return helpers.HandleSuccess(c, fiber.StatusOK, "Skill questions fetched successfully", remainingQuestions)
}

func GetBonusQuestions(c *fiber.Ctx) error {
	db := database.DB
	var questions []models.Question
	var answeredQuestionIDs []int
	var remainingQuestions []models.Question

	// Get userID from the context
	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing user_id", nil)
	}

	// Parse userID
	userID, err := uuid.Parse(userId)
	if err != nil {
		return helpers.HandleError(c, fiber.StatusBadRequest, "Invalid user ID format", err)
	}

	// Fetch today's 2 bonus questions
	err = db.Table("questions").
		Select("question_id, question_text, options, correct_answer, difficulty, points, multiplier, question_type, created_at").
		Where("question_type = ?", "Bonus").
		Order("created_at DESC").
		Limit(2).
		Find(&questions).Error

	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch bonus questions", err)
	}

	// Extract question IDs for today's bonus questions
	var bonusQuestionIDs []int
	for _, question := range questions {
		bonusQuestionIDs = append(bonusQuestionIDs, question.QuestionID)
	}

	// Check for questions already answered by the user
	err = db.Table("quiz_attempts").
		Select("question_id").
		Where("user_id = ?", userID).
		Where("question_id IN ?", bonusQuestionIDs). // Filter only today's bonus questions
		Find(&answeredQuestionIDs).Error

	if err != nil {
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch answered bonus questions", err)
	}

	// Filter out the answered questions from bonus questions
	answeredQuestionsMap := make(map[int]bool)
	for _, id := range answeredQuestionIDs {
		answeredQuestionsMap[id] = true
	}

	// Add only unanswered questions to remainingQuestions
	for _, question := range questions {
		if !answeredQuestionsMap[question.QuestionID] {
			remainingQuestions = append(remainingQuestions, question)
		}
	}

	// Return remaining questions (those not answered yet)
	return helpers.HandleSuccess(c, fiber.StatusOK, "Bonus questions fetched successfully", remainingQuestions)
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
