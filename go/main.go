package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const DATABASE_URL = "postgresql://order:orderpassword@localhost:5432/order"

type Post struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	Likes           int       `json:"likes"`
	Comments        int       `json:"comments"`
	Shares          int       `json:"shares"`
	EngagementScore float64   `json:"engagement_score,omitempty"`
}

type CreatePostRequest struct {
	Content string `json:"content" binding:"required"`
}

var dbPool *pgxpool.Pool

// Initialize connection pool
func initDB() {
	var err error
	dbPool, err = pgxpool.Connect(context.Background(), DATABASE_URL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	// Set max open connections, mimicking asyncpg default behavior
	dbPool.Config().MaxConns = 50
}

// Close the connection pool
func closeDB() {
	dbPool.Close()
}

// Calculate engagement score for a post
func calculateScore(post Post) float64 {
	timeDecay := time.Since(post.CreatedAt).Hours()
	return (2 * float64(post.Likes)) + (3 * float64(post.Comments)) + (5 * float64(post.Shares)) - timeDecay
}

// Create and fetch the top posts
func createAndFetchTopPosts(c *gin.Context) {
	var request CreatePostRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Fetch the last user ID
	var lastUserID int
	err := dbPool.QueryRow(context.Background(), "SELECT user_id FROM posts ORDER BY id DESC LIMIT 1").Scan(&lastUserID)
	if err != nil && err != pgx.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}
	userID := lastUserID + 1

	// Insert the new post
	createdAt := time.Now().UTC().Add(-time.Hour * time.Duration(rand.Intn(720))) // Random time within last 30 days
	likes := rand.Intn(101)
	comments := rand.Intn(51)
	shares := rand.Intn(21)

	var postID int
	err = dbPool.QueryRow(context.Background(),
		"INSERT INTO posts (user_id, content, created_at, likes, comments, shares) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		userID, request.Content, createdAt, likes, comments, shares).Scan(&postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	// Fetch the top 100,000 posts (this will be slow, adjust for larger datasets)
	rows, err := dbPool.Query(context.Background(), `
		SELECT id, user_id, content, created_at, likes, comments, shares
		FROM posts
		ORDER BY created_at DESC
		LIMIT 25000`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.Likes, &post.Comments, &post.Shares)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning row: " + err.Error()})
			return
		}
		post.EngagementScore = calculateScore(post)
		posts = append(posts, post)
	}

	// Sort posts by engagement score
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].EngagementScore > posts[j].EngagementScore
	})

	// Return top 10 posts
	if len(posts) > 10 {
		posts = posts[:10]
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Post created successfully",
		"post_id":   postID,
		"top_posts": posts,
	})
}

func main() {
	// Initialize the database connection pool
	initDB()
	defer closeDB()

	// Initialize the Gin router
	r := gin.Default()

	// Define routes
	r.POST("/create-and-fetch", createAndFetchTopPosts)

	// Start the server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
