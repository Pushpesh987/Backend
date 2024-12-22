package feed

import (
	"Backend/src/core/database"
	"Backend/src/core/helpers"
	"Backend/src/core/models"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeedPost struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Username        string    `json:"username"`
	ProfilePicURL   string    `json:"profile_pic_url"`
	Content         string    `json:"content"`
	MediaURL        string    `json:"media_url"`
	LikesCount      int       `json:"likes_count"`
	CommentsCount   int       `json:"comments_count"`
	Tags            []string  `json:"tags"`
	CreatedAt       time.Time `json:"created_at"`
	PopularityScore float64   `json:"popularity_score"`
}


func FetchFeed(c *fiber.Ctx) error {
	userId, ok := c.Locals("user_id").(string)
	if !ok || userId == "" {
		log.Println("Invalid or missing userID")
		return helpers.HandleError(c, fiber.StatusUnauthorized, "Invalid or missing auth_id", nil)
	}
	log.Printf("userID from JWT: %s\n", userId)

	userID, err := uuid.Parse(userId)
	if err != nil {
		log.Printf("Error fetching user: %v\n", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return helpers.HandleError(c, fiber.StatusNotFound, "User not found", nil)
		}
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Database query failed", err)
	}
	log.Printf("Fetched user ID: %s\n", userID)

	limit, offset := ParsePagination(c)
	log.Printf("Pagination parameters - limit: %d, offset: %d\n", limit, offset)

	connections, err := GetUserConnections(userID)
	if err != nil {
		log.Printf("Error fetching user connections: %v\n", err)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch user connections", err)
	}
	log.Printf("Fetched connections: %v\n", connections)

	excludedPosts, err := GetLikedPostIDs(userID)
	if err != nil {
		log.Printf("Error fetching liked posts: %v\n", err)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch excluded posts", err)
	}
	log.Printf("Fetched excluded post IDs: %v\n", excludedPosts)

	posts, err := GetEnhancedFeedPosts(userID, connections, excludedPosts, limit, offset)
	if err != nil {
		log.Printf("Error fetching posts: %v\n", err)
		return helpers.HandleError(c, fiber.StatusInternalServerError, "Failed to fetch feed", err)
	}
	log.Printf("Fetched %d posts: %+v\n", len(posts), posts)

	feedPosts := CalculatePopularityAndRetrieveTags(posts)
	log.Printf("Enhanced feed posts: %+v\n", feedPosts)

	SortByPopularity(feedPosts)
	log.Printf("Sorted feed posts by popularity")

	return helpers.HandleSuccess(c, fiber.StatusOK, "Feed fetched successfully", feedPosts)
}

func GetUserConnections(userID uuid.UUID) ([]string, error) {
	db := database.DB
	log.Printf("Fetching connections for userID: %s\n", userID)
	var connections []struct {
		ConnectionID string `gorm:"column:connection_id"`
	}
	if err := db.Table("connections").Where("user_id = ?", userID).Select("connection_id").Find(&connections).Error; err != nil {
		return nil, err
	}
	connectionIDs := make([]string, len(connections))
	for i, conn := range connections {
		connectionIDs[i] = conn.ConnectionID
	}
	log.Printf("Fetched connections: %v\n", connectionIDs)
	return connectionIDs, nil
}

func GetLikedPostIDs(userID uuid.UUID) ([]string, error) {
	db := database.DB
	log.Printf("Fetching liked post IDs for userID: %s\n", userID)
	var likedPosts []struct {
		PostID string `gorm:"column:post_id"`
	}
	if err := db.Table("likes").Where("user_id = ?", userID).Select("post_id").Find(&likedPosts).Error; err != nil {
		return nil, err
	}
	postIDs := make([]string, len(likedPosts))
	for i, post := range likedPosts {
		postIDs[i] = post.PostID
	}
	log.Printf("Fetched liked post IDs: %v\n", postIDs)
	return postIDs, nil
}

func GetEnhancedFeedPosts(userID uuid.UUID, connections, excludedPosts []string, limit, offset int) ([]models.Post, error) {
	db := database.DB
	log.Printf("Fetching filtered posts with userID: %s, connections: %v, excludedPosts: %v, limit: %d, offset: %d\n", userID, connections, excludedPosts, limit, offset)

	var posts []models.Post

	// Fetch user tags based on their posts
	var userTags []string
	if err := db.Table("post_tags").
		Joins("JOIN tags ON post_tags.tag_id = tags.id").
		Where("post_tags.post_id IN (?)", db.Table("posts").Select("id").Where("user_id = ?", userID)).
		Select("DISTINCT tags.tag").
		Find(&userTags).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user tags: %w", err)
	}

	// Fetch user interests
	var userInterests []string
	if err := db.Table("user_interests").
		Joins("JOIN interests ON user_interests.interest_id = interests.interest_id").
		Where("user_interests.user_id = ?", userID).
		Select("DISTINCT interests.interest_name").
		Find(&userInterests).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user interests: %w", err)
	}

	// Combine tags and interests for matching posts
	combinedTagsAndInterests := append(userTags, userInterests...)
	log.Printf("Combined tags and interests: %v", combinedTagsAndInterests)

	// Fetch most liked posts
	var mostLikedPostIDs []string
	if err := db.Table("likes").
		Select("post_id").
		Group("post_id").
		Order("COUNT(post_id) DESC").
		Limit(3).
		Find(&mostLikedPostIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch most liked posts: %w", err)
	}

	// Fetch most commented posts
	var mostCommentedPostIDs []string
	if err := db.Table("comments").
		Select("post_id").
		Group("post_id").
		Order("COUNT(post_id) DESC").
		Limit(3).
		Find(&mostCommentedPostIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch most commented posts: %w", err)
	}

	// Combine liked and commented posts for weighted ranking
	weightedPostIDs := append(mostLikedPostIDs, mostCommentedPostIDs...)

	// Query for posts from connections, now including username and profile_pic_url
	query := db.Table("posts").
		Select(`posts.id, posts.user_id, posts.content, posts.media_url, 
                posts.likes_count, posts.comments_count, posts.created_at, 
                users.username, users.profile_pic_url`).
		Joins("JOIN users ON posts.user_id = users.id").
		Where("posts.user_id IN (?)", connections)

	// Exclude certain posts
	if len(excludedPosts) > 0 {
		query = query.Where("posts.id NOT IN (?)", excludedPosts)
	}

	// Fetch posts from connections
	var connectionPosts []models.Post
	if err := query.Order("posts.created_at DESC").Limit(limit).Offset(offset).Find(&connectionPosts).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch posts from connections: %w", err)
	}

	// Query for posts matching user tags and interests
	var tagsAndInterestsPosts []models.Post
	if len(combinedTagsAndInterests) > 0 {
		tagsAndInterestsQuery := db.Table("posts").
			Select(`posts.id, posts.user_id, posts.content, posts.media_url, 
                    posts.likes_count, posts.comments_count, posts.created_at, 
                    users.username, users.profile_pic_url`).
			Where("posts.id NOT IN (?)", excludedPosts).
			Where("posts.id IN (?)",
				db.Table("post_tags").
					Joins("JOIN tags ON post_tags.tag_id = tags.id").
					Where("tags.tag IN (?)", combinedTagsAndInterests).
					Select("post_tags.post_id")).
			Joins("JOIN users ON posts.user_id = users.id").
			Order("posts.created_at DESC").
			Limit(limit).
			Offset(offset)

		if err := tagsAndInterestsQuery.Find(&tagsAndInterestsPosts).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch posts matching tags: %w", err)
		}
	}

	// Query for posts based on weighted IDs (most liked and most commented)
	var weightedPosts []models.Post
	if len(weightedPostIDs) > 0 {
		weightedQuery := db.Table("posts").
			Select(`posts.id, posts.user_id, posts.content, posts.media_url, 
                    posts.likes_count, posts.comments_count, posts.created_at, 
                    users.username, users.profile_pic_url`).
			Where("posts.id IN (?)", weightedPostIDs).
			Joins("JOIN users ON posts.user_id = users.id").
			Order("posts.created_at DESC").
			Limit(limit).
			Offset(offset)

		if err := weightedQuery.Find(&weightedPosts).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch weighted posts: %w", err)
		}
	}

	// Combine all post results
	posts = append(posts, connectionPosts...)
	posts = append(posts, tagsAndInterestsPosts...)
	posts = append(posts, weightedPosts...)

	// Deduplicate posts based on ID (avoid duplication across categories)
	posts = deduplicatePosts(posts)
	log.Printf("Retrieved filtered posts: %+v\n", posts)

	return posts, nil
}


func deduplicatePosts(posts []models.Post) []models.Post {
	seen := make(map[string]bool)
	uniquePosts := []models.Post{}

	for _, post := range posts {
		idStr := post.ID.String()
		if !seen[idStr] {
			seen[idStr] = true
			uniquePosts = append(uniquePosts, post)
		}
	}

	return uniquePosts
}

func CalculatePopularityAndRetrieveTags(posts []models.Post) []FeedPost {
	log.Println("Calculating popularity scores and retrieving tags")
	feedPosts := make([]FeedPost, len(posts))

	for i, post := range posts {
		tags, err := RetrieveTagsForPost(post.ID.String())
		if err != nil {
			log.Printf("Error retrieving tags for post ID %s: %v\n", post.ID.String(), err)
		}
		if len(tags) == 0 {
			tags = []string{}
		}

		score := CalculateScore(post.LikesCount, post.CommentsCount, post.CreatedAt)
		log.Printf("Post ID: %s, Score: %.2f\n", post.ID.String(), score)

		feedPosts[i] = FeedPost{
			ID:              post.ID.String(),
			UserID:          post.UserID.String(),
			Username:        post.Username,
			ProfilePicURL:   post.ProfilePicURL,
			Content:         post.Content,
			MediaURL:        post.MediaURL,
			LikesCount:      post.LikesCount,
			CommentsCount:   post.CommentsCount,
			Tags:            tags,
			CreatedAt:       post.CreatedAt,
			PopularityScore: score,
		}
	}
	return feedPosts
}

func RetrieveTagsForPost(postID string) ([]string, error) {
	db := database.DB
	log.Printf("Retrieving tags for post ID: %s\n", postID)
	var tags []string
	query := `
		SELECT t.tag
		FROM tags t
		JOIN post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = ?;
	`
	if err := db.Raw(query, postID).Scan(&tags).Error; err != nil {
		return nil, fmt.Errorf("error retrieving tags for post ID %v: %w", postID, err)
	}
	log.Printf("Retrieved tags for post ID %s: %v\n", postID, tags)
	return tags, nil
}

func CalculateScore(likes, comments int, createdAt time.Time) float64 {
	daysSincePost := time.Since(createdAt).Hours() / 24
	if daysSincePost <= 0 {
		daysSincePost = 1
	}
	score := float64(likes*2+comments*3) / daysSincePost
	log.Printf("Calculated score: %.2f (likes: %d, comments: %d, days since post: %.2f)\n", score, likes, comments, daysSincePost)
	return score
}

func SortByPopularity(posts []FeedPost) {
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].PopularityScore > posts[j].PopularityScore
	})
}

func ParsePagination(c *fiber.Ctx) (int, int) {
	limit, err := strconv.Atoi(c.Query("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	return limit, offset
}
