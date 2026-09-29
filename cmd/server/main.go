package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var API_URL string
var API_KEY string

func main() {

	// 環境変数の読み込み
	err := godotenv.Load()
	if err != nil {
		panic("環境変数が読み込めませんでした")
	}

	API_URL = os.Getenv("SUPABASE_URL")
	API_KEY = os.Getenv("SUPABASE_KEY")

	router := gin.Default() // Ginルーター初期化

	router.POST("/api/today", uploadImage) //今日の画像を投稿
	router.GET("/api/today", getImage)     //今日の画像を表示
	// router.GET("/api/todayflag", getTodayflag) //今日投稿したかどうか

	// webフォルダをFrontendとして配信
	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	router.Static("/js", "./web/js") // javascriptのフォルダ配信

	// 標準出力にメッセージ表示
	fmt.Println("Server started: http://localhost:3000")

	//サーバ起動
	router.Run(":3000")

}

// 画像を投稿する関数
func uploadImage(c *gin.Context) {
	// フロントから画像を受け取る
	image, err := c.FormFile("test")
	errlog(err)

	//画像のContent-Type(jpeg/pngとか)を取得
	conType := image.Header.Get("Content-Type")
	if conType != "image/jpeg" && conType != "image/png" {
		log.Println("jpg、png以外のファイルは受け取れません")
		return
	}

	// 画像を開く
	imageIO, err := image.Open()
	errlog(err)
	defer imageIO.Close()
	//画像名を日付で作成
	imageName := "today-" + time.Now().Format("20060102150405") + filepath.Ext(image.Filename)

	// // 一時ファイルとして保存
	// tempFile, err := os.CreateTemp("", "temp-*.jpeg")
	// errlog(err)
	// defer os.Remove(tempFile.Name()) // 終了時に一時ファイルを削除
	// defer tempFile.Close()
	// err = c.SaveUploadedFile(image, tempFile.Name())
	// errlog(err)

	// Supabase Storageに画像をアップロード

	storageURL := API_URL + "/storage/v1/object/images/" + imageName // StorageのアップロードURL

	// HTTPリクエストを作成
	req, err := http.NewRequest("POST", storageURL, imageIO)
	errlog(err)
	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("apikey", API_KEY)
	req.Header.Set("Content-Type", conType)

	// HTTPリクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	errlog(err)
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Println("画像のアップロードに失敗しました。ステータスコード:", resp.StatusCode)
		return
	} else {
		log.Printf("画像をアップロードできました！「%s」\n", imageName)

		// 投稿できたら、データベースに投稿したことを記録する
		// 挿入するデータを作成
		post := struct {
			ImagePath string `json:"image_path"`
		}{
			ImagePath: imageName,
		}
		// JSONに変換
		postJSON, err := json.Marshal(post)
		errlog(err)

		// storageと同様に通信
		// HTTPリクエストを作成
		req, err := http.NewRequest("POST", API_URL+"/rest/v1/posts", bytes.NewReader(postJSON))
		errlog(err)
		req.Header.Set("Authorization", "Bearer "+API_KEY)
		req.Header.Set("apikey", API_KEY)
		req.Header.Set("Content-Type", "application/json")

		// HTTPリクエストを送信
		client := &http.Client{}
		resp, err := client.Do(req)
		errlog(err)
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Println("投稿に失敗しました。ステータスコード:", resp.StatusCode)
			return
		} else {
			log.Printf("投稿できました！「%s」\n", post.ImagePath)
		}
	}

}

// 今日の画像そのものを返す
func getImage(c *gin.Context) {

	imageName := "today.jpeg"
	imagePath := filepath.Join("upload", filepath.Base(imageName))

	_, err := os.Stat(imagePath)
	errlog(err)

	c.File(imagePath)

}

// エラーログを出力する関数
func errlog(err error) {
	if err != nil {
		log.Println(err)
		return
	}
}
