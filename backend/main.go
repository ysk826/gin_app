package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Ginルーターの初期化
	r := gin.Default()

	// CORSミドルウェアの設定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// apiエンドポイントのグループ化
	api := r.Group("/api")
	{
		api.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Hello,燈哉!",
			})
		})
	}

	// 日記のエンドポイント
	api.GET("/diary", getDiaryEntriesByDate)
	api.POST("/diary", createDiaryEntry)
	//api.PUT("/diary/:id", updateDiaryEntry)

	// サーバーの起動（エラーハンドリングを追加）
	log.Println("サーバーを開始します...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("サーバー起動エラー: %v", err)
	}
}
