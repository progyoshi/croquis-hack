package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

//今日画像が投稿されたかどうか
// func getTodayflag(c *gin.Context){
// 	//etc...
// }

// 画像を受け取るテスト
func uploadImage(c *gin.Context) {
	//画像を受け取る
	file, err := c.FormFile("test")
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "画像を取得できませんでした",
		})
		return
	}

	//画像保存先
	imagePath := filepath.Join("upload", "today")
	//保存
	err = c.SaveUploadedFile(file, imagePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "画像を保存できませんでした",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "画像をアップロードしました",
	})
}

// 今日の画像そのものを返す
func getImage(c *gin.Context) {

	imagePath := filepath.Join("upload", "today")

	_, err := os.Stat(imagePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "画像がありません",
		})
		return
	}

	c.File(imagePath)

}

func main() {
	// http.HandleFunc("/", testHandler)

	router := gin.Default() // Ginルーター初期化

	router.POST("/api/today", uploadImage) //今日の画像を投稿
	router.GET("/api/today", getImage)     //今日の画像を表示
	// router.GET("/api/todayflag", getTodayflag) //今日投稿したかどうか

	// webフォルダをFrontendとして配信
	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	router.StaticFile("/js/api.js", "./web/js/api.js")

	// 標準出力にメッセージ表示
	fmt.Println("Server started: http://localhost:3000")

	//サーバ起動
	router.Run(":3000")

}
