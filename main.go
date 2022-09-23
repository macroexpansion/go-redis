package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"goredis/db"
	"net/http"
	"time"
)

func test() {
	opt, err := redis.ParseURL("redis://localhost:6379/")
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(opt)

	ctx := context.Background()
	rdb.Set(ctx, "key", "value", 0)

	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println(val)
}

var (
	ListenAddr = "localhost:8080"
	RedisAddr  = "localhost:6379"
)

func initRouter(database *db.Database) *gin.Engine {
	r := gin.Default()

	r.POST("/points", func(c *gin.Context) {
		var userJson db.User
		if err := c.ShouldBindJSON(&userJson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := database.SaveUser(&userJson)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": userJson})
	})

	r.GET("/points/:username", func(c *gin.Context) {
		username := c.Param("username")
		user, err := database.GetUser(username)
		if err != nil {
			if err == db.ErrNil {
				c.JSON(http.StatusNotFound, gin.H{"error": "No record found for " + username})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": user})
	})

	r.GET("/points", func(c *gin.Context) {
		leaderboard, err := database.GetLeaderboard()
		if err != nil {
			if err == db.ErrNil {
				c.JSON(http.StatusNotFound, gin.H{"error": "No record found for leaderboard"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": leaderboard})
	})

	return r
}

func testRedisTxPipeline(database *db.Database) {
	pipe := database.Client.TxPipeline()
	pipe.Set(db.Ctx, "key", "value", 1*time.Second)
	pipe.Set(db.Ctx, "num", 500, 1*time.Second)
	results, _ := pipe.Exec(db.Ctx)
	for _, value := range results {
		fmt.Println(value)
	}
}

func main() {
	database, err := db.NewDatabase(RedisAddr)
	if err != nil {
		panic(err)
	}
	// testRedisTxPipeline(database)

	router := initRouter(database)
	router.Run(ListenAddr)
}
