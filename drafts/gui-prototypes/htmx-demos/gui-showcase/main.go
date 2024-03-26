
package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
  "fmt"
)

func main() {
	r := gin.Default()

	// Serve the HTML file
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Echo back what's sent to it
	r.POST("/echo", func(c *gin.Context) {
		value := c.PostForm("value")
		c.String(http.StatusOK, value)
	})

	// Handle radio button selection
	r.GET("/radio", func(c *gin.Context) {
		option := c.Query("radio")
		c.JSON(http.StatusOK, gin.H{"selected": option})
	})

	// Simulate file upload response
	r.POST("/upload", func(c *gin.Context) {
		// In a real application, you would handle the uploaded file here
		time.Sleep(1 * time.Second) // Simulate some processing time
		c.JSON(http.StatusOK, gin.H{"status": "File uploaded successfully"})
	})

	// Provide data for a button click
	r.GET("/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "You clicked the button!"})
	})

	// Generate table data
	r.GET("/table", func(c *gin.Context) {
		c.HTML(http.StatusOK, "partial_table.html", gin.H{
			"Rows": []gin.H{
				{"Column1": "Row 1 Col 1", "Column2": "Row 1 Col 2"},
				{"Column1": "Row 2 Col 1", "Column2": "Row 2 Col 2"},
			},
		})
	})

	// Handle the min/max range request
	r.GET("/range", func(c *gin.Context) {
		minPrice := c.Query("minPrice")
		maxPrice := c.Query("maxPrice")
		response := fmt.Sprintf("Selected range: $%s - $%s", minPrice, maxPrice)
		c.String(http.StatusOK, response)
	})

	r.Run() // Listen and serve on 0.0.0.0:8080
}
