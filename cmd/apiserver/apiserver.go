package main

import (
	"strconv"

	"github.com/Xinrea/ffreplay/internal/data/markers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/markers/:code/:id", func(c *gin.Context) {
		code := c.Param("code")

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(401, gin.H{
				"message": "invalid fight id",
			})

			return
		}

		markers := markers.QueryWorldMarkers(code, id)
		c.JSON(200, gin.H{
			"data": markers,
		})
	})
	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	r.Run()
}
