package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/akamensky/argparse"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/geulgyeol/link-kv/db"
	"github.com/geulgyeol/link-kv/local"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	parser := argparse.NewParser("geulgyeol-link-kv", "A HTML storage server for Geulgyeol.")

	port := parser.Int("p", "port", &argparse.Options{Default: 8080, Help: "Port to run the server on"})
	connString := parser.String("c", "conn-string", &argparse.Options{
		Help: "PostgreSQL connection string",
	})
	localMode := parser.Flag("l", "local", &argparse.Options{Help: "Whether to use emulator"})

	err := parser.Parse(os.Args)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	var queries local.LinkKVService

	if *localMode {
		queries = local.New()
		fmt.Println("Running in local emulation mode (in-memory storage)")
	} else {
		if *connString == "" {
			panic("conn-string is required when not in local mode")
		}

		pool, err := pgxpool.New(ctx, *connString)
		if err != nil {
			panic(fmt.Sprintf("Unable to connect to database: %v", err))
		}
		defer pool.Close()

		queries = db.New(pool)
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// =========
	// Users (CRUD)
	// =========

	r.GET("/users", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if limit <= 0 || limit > 1000 {
			limit = 50
		}

		users, err := queries.ListBlogUsers(c.Request.Context(), int32(limit))
		if err != nil {
			fmt.Printf("Error listing users: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, users)
	})

	r.GET("/users/:platform/:user_id", func(c *gin.Context) {
		user, err := queries.GetBlogUser(c.Request.Context(), db.GetBlogUserParams{
			BlogPlatform: c.Param("platform"),
			UserID:       c.Param("user_id"),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(404, gin.H{"error": "Not found"})
				return
			}
			fmt.Printf("Error getting user: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, user)
	})

	r.POST("/users", func(c *gin.Context) {
		var body struct {
			BlogPlatform string `json:"blog_platform" binding:"required"`
			UserID       string `json:"user_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{"error": "blog_platform and user_id are required"})
			return
		}

		user, err := queries.CreateBlogUser(c.Request.Context(), db.CreateBlogUserParams{
			BlogPlatform: body.BlogPlatform,
			UserID:       body.UserID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(409, gin.H{"error": "Already exists"})
				return
			}
			fmt.Printf("Error creating user: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(201, user)
	})

	r.PUT("/users/:platform/:user_id", func(c *gin.Context) {
		params := db.UpdateBlogUserLastEnqueuedAtParams{
			BlogPlatform: c.Param("platform"),
			UserID:       c.Param("user_id"),
		}

		// Verify user exists first
		_, err := queries.GetBlogUser(c.Request.Context(), db.GetBlogUserParams{
			BlogPlatform: params.BlogPlatform,
			UserID:       params.UserID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(404, gin.H{"error": "Not found"})
				return
			}
			fmt.Printf("Error getting user: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if err := queries.UpdateBlogUserLastEnqueuedAt(c.Request.Context(), params); err != nil {
			fmt.Printf("Error updating user: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, gin.H{"status": "updated"})
	})

	r.DELETE("/users/:platform/:user_id", func(c *gin.Context) {
		err := queries.DeleteBlogUser(c.Request.Context(), db.DeleteBlogUserParams{
			BlogPlatform: c.Param("platform"),
			UserID:       c.Param("user_id"),
		})
		if err != nil {
			fmt.Printf("Error deleting user: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, gin.H{"status": "deleted"})
	})

	// =========
	// Posts (Read-only)
	// =========

	r.GET("/posts", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if limit <= 0 || limit > 1000 {
			limit = 50
		}

		platform := c.Query("platform")

		if platform != "" {
			posts, err := queries.ListBlogPostsByPlatform(c.Request.Context(), db.ListBlogPostsByPlatformParams{
				BlogPlatform: platform,
				Limit:        int32(limit),
			})
			if err != nil {
				fmt.Printf("Error listing posts: %v\n", err)
				c.JSON(500, gin.H{"error": "Internal server error"})
				return
			}
			c.JSON(200, posts)
			return
		}

		posts, err := queries.ListBlogPosts(c.Request.Context(), int32(limit))
		if err != nil {
			fmt.Printf("Error listing posts: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, posts)
	})

	r.GET("/posts/:platform/*post_url", func(c *gin.Context) {
		platform := c.Param("platform")
		postURL := c.Param("post_url")
		if len(postURL) > 0 && postURL[0] == '/' {
			postURL = postURL[1:]
		}
		postURL, err := url.PathUnescape(postURL)
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid post URL"})
			return
		}

		post, err := queries.GetBlogPost(c.Request.Context(), postURL)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(404, gin.H{"error": "Not found"})
				return
			}
			fmt.Printf("Error getting post: %v\n", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if post.BlogPlatform != platform {
			c.JSON(404, gin.H{"error": "Not found"})
			return
		}

		c.JSON(200, post)
	})

	fmt.Printf("Starting server on port %d\n", *port)

	// run the server
	_ = r.Run(fmt.Sprintf(":%d", *port))
}
