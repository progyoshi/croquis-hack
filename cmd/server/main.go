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
	// この処理超出てくるよ、エラーがあった時ログを出す＆ここで関数を中断するよ
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "画像の受け取りに失敗しました",
		})
		return
	}

	//画像のContent-Type(jpeg/pngとか)を取得
	conType := image.Header.Get("Content-Type")
	if conType != "image/jpeg" && conType != "image/png" {
		log.Println("jpg、png以外のファイルは受け取れません")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "jpg、png以外のファイルは受け取れません",
		})
		return
	}

	// 画像を開く
	imageIO, err := image.Open()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "画像のオープンに失敗しました",
		})
		return
	}
	defer imageIO.Close()
	//画像名を日付で作成
	imageName := "today-" + time.Now().Format("20060102150405") + filepath.Ext(image.Filename)

	// Supabase Storageに画像をアップロード
	storageURL := API_URL + "/storage/v1/object/images/" + imageName // StorageのアップロードURL

	// HTTPリクエストを作成
	req, err := http.NewRequest("POST", storageURL, imageIO)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("apikey", API_KEY)
	req.Header.Set("Content-Type", conType)

	// HTTPリクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Println("画像のアップロードに失敗しました。ステータスコード:", resp.StatusCode)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "画像のアップロードに失敗しました",
		})
		return
	}
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
	if err != nil {
		log.Println(err)
		return
	}

	// storageと同様に通信
	// HTTPリクエストを作成
	req, err = http.NewRequest("POST", API_URL+"/rest/v1/posts", bytes.NewReader(postJSON))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("apikey", API_KEY)
	req.Header.Set("Content-Type", "application/json")

	// HTTPリクエストを送信
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "投稿に失敗しました",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Println("投稿に失敗しました。ステータスコード:", resp.StatusCode)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "投稿に失敗しました",
		})
		return
	}
	log.Printf("投稿できました！「%s」\n", post.ImagePath)

	// ストレージ上の画像のパスを返す
	imageURL := API_URL + "/storage/v1/object/public/images/" + post.ImagePath
	log.Printf("画像のURLを返します:%s\n", imageURL)

	// これが画像投稿した時に画像のURL返すJSONだよ
	c.JSON(http.StatusOK, gin.H{
		"message":   "投稿に成功しました",
		"image_url": imageURL, // 画像のURLを返す
	})
}

// 今日の画像のURLを返す
func getImage(c *gin.Context) {

	// 現在の日本時刻を取得
	jst, err := time.LoadLocation("Asia/Tokyo") // 日本時間のタイムゾーンを取得
	if err != nil {
		log.Println(err)
		return
	}
	// 4:00～3:59を1日とするため、現在時刻から4時間引いた時刻を取得
	now := time.Now().In(jst).Add(time.Duration(-4) * time.Hour)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, jst)
	endOfDay := startOfDay.AddDate(0, 0, 1)

	log.Printf("now:%s, start:%s, end:%s", now.String(), startOfDay.String(), endOfDay.String())
	log.Printf("UTC start: %s, end: %s", startOfDay.UTC(), endOfDay.UTC())

	// PostgREST APIのクエリを作成
	// 4:00～翌3:59の投稿を昇順に並べ、上位1つを取得
	url := API_URL + "/rest/v1/posts?select=image_path,posted_at" + "&posted_at=gte." + startOfDay.UTC().Format(time.RFC3339) + "&posted_at=lt." + endOfDay.UTC().Format(time.RFC3339) + "&order=posted_at.desc&limit=1"

	// HTTPリクエストを作成
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("apikey", API_KEY)

	// HTTPリクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("投稿の取得に失敗しました。ステータスコード:", resp.StatusCode)
		// ここでも一応フロントに返してるよ：500
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "投稿の取得に失敗しました。",
		})
		return
	}

	// レスポンスをデコード
	var posts []struct {
		ImagePath string `json:"image_path"`
		PostedAt  string `json:"posted_at"`
	}

	err = json.NewDecoder(resp.Body).Decode(&posts)
	if err != nil {
		log.Println(err)
		return
	}

	if len(posts) == 0 {
		log.Println("今日の投稿が見つかりませんでした。")
		// ここでも一応フロントに返してるよ：404
		c.JSON(http.StatusNotFound, gin.H{
			"message": "今日の投稿が見つかりませんでした。",
		})
		return
	}

	log.Printf("投稿を取得できました！:%+v\n", posts[0])

	// ストレージ上の画像のパスを返す
	imageURL := API_URL + "/storage/v1/object/public/images/" + posts[0].ImagePath
	log.Printf("画像のURLを返します:%s\n", imageURL)

	// これが投稿した画像のJSONだよ
	c.JSON(http.StatusOK, gin.H{
		"image_url": imageURL, // 画像のURL
		// "posted_at": posts[0].PostedAt, // UTC時間での投稿日時返さなくても良いかなって
	})
}
