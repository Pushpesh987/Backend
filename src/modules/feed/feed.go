package feed

import (
	"Backend/src/core/database"
	"Backend/src/core/models"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"

	// "math/rand"
	"net/http"
	// "sort"
	// "time"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func fetchAllPosts() ([]models.Post, error) {
	var posts []models.Post
	if err := database.DB.Find(&posts).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	return posts, nil
}

func fetchRecommendedPosts(userID string, posts []models.Post) ([]string, error) {
	url := "http://localhost:5000/recommend"

	payload := map[string]interface{}{
		"user_id": userID,
		"posts":   posts,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var recommendedPostIDs []string
	if err := json.NewDecoder(resp.Body).Decode(&recommendedPostIDs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return recommendedPostIDs, nil
}

func GetFeed(c *fiber.Ctx) error {
	userID := c.Params("user_id")

	// Fetch recommendations using the helper function
	recommendedPosts, err := fetchRecommendations(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "Failed to fetch feed",
			"detail": err.Error(),
		})
	}

	// Return the recommended posts as a response
	return c.JSON(fiber.Map{
		"recommended_posts": recommendedPosts,
	})
}

// func FeedHandler(c *fiber.Ctx) error {
// 	// Extract user ID from the context (e.g., from JWT or query params)
// 	userID := c.Locals("user_id").(string) // Adjust based on your JWT setup

// 	// Call GetFeed to get the recommended posts
// 	recommendedPosts, err := GetFeed(userID)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch feed",
// 			"detail": err.Error(),
// 		})
// 	}

// 	// Return the recommended posts as JSON
// 	return c.JSON(recommendedPosts)
// }

func fetchRecommendations(userID string) ([]string, error) {
	apiURL := "http://127.0.0.1:5000/recommend"

	// Prepare the payload
	payload := map[string]interface{}{
		"user_id": userID,
		"posts": []map[string]interface{}{
			{"id": "1"},
			{"id": "2"},
			{"id": "3"},
		},
	}

	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("failed to marshal payload")
	}

	// Send POST request to Python service
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, errors.New("failed to call recommendation service")
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read response body")
	}

	// Decode JSON response
	var recommendedPosts []string
	err = json.Unmarshal(respBody, &recommendedPosts)
	if err != nil {
		return nil, errors.New("failed to decode recommendations")
	}

	return recommendedPosts, nil
}