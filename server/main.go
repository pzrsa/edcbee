package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var APIv1 *gin.RouterGroup

func main() {
	router := gin.New()

	APIv1 = router.Group("api/v1")

	// API Version 1
	GetStatus(APIv1)

	router.Run(":8080")
}

func GetStatus(router *gin.RouterGroup) {
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
