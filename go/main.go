package main

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

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

func calculateScore(post Post) float64 {
	timeDecay := time.Since(post.CreatedAt).Hours()
	return (2 * float64(post.Likes)) + (3 * float64(post.Comments)) + (5 * float64(post.Shares)) - timeDecay
}

func getTopPosts(c *gin.Context) {
	connStr := "host=localhost port=5432 user=order password=orderpassword dbname=order sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, user_id, content, created_at, likes, comments, shares FROM posts LIMIT 100")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var posts []Post = make([]Post, 0)
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.Likes, &post.Comments, &post.Shares)
		if err != nil {
			log.Fatal(err)
		}
		post.EngagementScore = calculateScore(post)
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].EngagementScore > posts[j].EngagementScore
	})

	if len(posts) < 10 {
		c.JSON(200, posts)
	}

	c.JSON(200, posts[:10])
}

func main() {
	r := gin.Default()

	r.GET("/top-posts", getTopPosts)

	if err := r.Run(":8080"); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
