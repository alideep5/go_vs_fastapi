package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})

	r.GET("/task", func(c *gin.Context) {
		time.Sleep(1 * time.Second)
		c.JSON(200, gin.H{"message": "I/O operation complete!"})
	})

	r.Run(":8080")
}
