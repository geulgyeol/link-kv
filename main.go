package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/akamensky/argparse"
	"github.com/cockroachdb/pebble"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	parser := argparse.NewParser("geulgyeol-link-kv", "A HTML storage server for Geulgyeol.")

	port := parser.Int("p", "port", &argparse.Options{Default: 8080, Help: "Port to run the server on"})
	dataPath := parser.String("d", "data-path", &argparse.Options{Default: "/data", Help: "Path to the Pebble database"})

	err := parser.Parse(os.Args)
	if err != nil {
		panic(err)
	}

	db, err := pebble.Open(*dataPath, &pebble.Options{})
	if err != nil {
		panic(err)
	}

	defer func(db *pebble.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("Error closing database: %v\n", err)
		}
	}(db)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	/* GET /{링크}
	-> 200 OK / 404 Not found

	POST /{링크}
	-> 201 Created / 409 Conflict
	*/

	r.GET("/:link", func(c *gin.Context) {
		link := c.Param("link")

		_, closer, err := db.Get([]byte(link))
		if err != nil {
			if errors.Is(err, pebble.ErrNotFound) {
				c.JSON(404, gin.H{"error": "Not found"})
				return
			}

			fmt.Printf("Error getting value from database: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		defer func(closer io.Closer) {
			err := closer.Close()
			if err != nil {
				fmt.Printf("Error closing value reader: %v\n", err)
			}
		}(closer)

		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/:link", func(c *gin.Context) {
		link := c.Param("link")

		_, closer, err := db.Get([]byte(link))
		if err == nil {
			_ = closer.Close()
			c.JSON(409, gin.H{"error": "Already exists"})
			return
		}

		if !errors.Is(err, pebble.ErrNotFound) {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		err = db.Set([]byte(link), []byte(""), pebble.Sync)
		if err != nil {
			fmt.Printf("Error setting value in database: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(201, gin.H{"status": "created"})
	})

	fmt.Printf("Starting server on port %d\n", *port)

	// run the server
	_ = r.Run(fmt.Sprintf(":%d", *port))
}
