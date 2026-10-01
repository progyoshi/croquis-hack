package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var API_URL string
var API_KEY string
var userID = 1 // デモユーザに固定してるよ

func main() {

	// 環境変数の読み込み
	err := godotenv.Load()
	if err != nil {
		panic("環境変数が読み込めませんでした")
	}
	API_URL = os.Getenv("SUPABASE_URL")
	API_KEY = os.Getenv("SUPABASE_KEY")
	if API_URL == "" || API_KEY == "" {
		panic("環境変数が設定されていません")
	}

	router := gin.Default()             // Ginルーター初期化
	router.MaxMultipartMemory = 8 << 20 // 8MBまでのファイルを受け付ける

	router.POST("/api/today", uploadImage) //今日の画像を投稿
	router.GET("/api/today", getImage)     //今日の画像を表示

	// webフォルダをFrontendとして配信
	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	router.Static("/js", "./web/js")   // javascriptのフォルダ配信
	router.Static("/img", "./web/img") // 画像のフォルダ配信

	// 標準出力にメッセージ表示
	fmt.Println("Server started: http://localhost:3000")

	//サーバ起動
	router.Run(":3000")

}

// 画像を投稿する関数
func uploadImage(c *gin.Context) {
	// 8MBまでのファイルを受け付ける
	const maxImageSize = 8 << 20 // 8MB
	// multipart/form-dataのヘッダーサイズを考慮したら1MB追加した方が良いらしい
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageSize+(1<<20))

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

	if image.Size > maxImageSize {
		log.Println("画像のサイズが大きすぎます")
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"message": "画像のサイズが大きすぎます",
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

	imageHeader := make([]byte, 512) // 画像のヘッダーを取得するためのバッファ
	n, err := imageIO.Read(imageHeader)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "画像の読み込みに失敗しました",
		})
		return
	}

	//画像のContent-Type(jpeg/pngとか)を取得
	conType := http.DetectContentType(imageHeader[:n])
	if conType != "image/jpeg" && conType != "image/png" {
		log.Println("jpg、png以外のファイルは受け取れません")
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"message": "jpg、png以外のファイルは受け取れません",
		})
		return
	}

	// ファイル位置を先頭に戻す
	_, err = imageIO.Seek(0, io.SeekStart)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "画像のシークに失敗しました",
		})
		return
	}

	// 拡張子を決定
	ext := ".jpg"
	if conType == "image/png" {
		ext = ".png"
	}
	//画像名をランダムにする
	randBytes := make([]byte, 8) // 8バイトのランダムなバイト列を生成
	_, err = rand.Read(randBytes)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "ランダムな画像名の生成に失敗しました",
		})
		return
	}
	imageName := hex.EncodeToString(randBytes) + ext

	// StorageのアップロードURL
	// 一旦ユーザID1に固定してるよ
	storageURL := API_URL + "/storage/v1/object/" + strconv.Itoa(userID) + "/images/" + imageName

	// Supabase Storageに画像をアップロード
	// HTTPリクエストを作成
	req, err := http.NewRequest("POST", storageURL, imageIO)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "HTTPリクエストの作成に失敗しました",
		})
		return
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("apikey", API_KEY)
	req.Header.Set("Content-Type", conType)

	// HTTPリクエストを送信
	client := &http.Client{
		Timeout: 10 * time.Second, // タイムアウトを設定
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "HTTPリクエストの送信に失敗しました",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Println("画像のアップロードに失敗しました。ステータスコード:", resp.StatusCode)
		c.JSON(http.StatusBadGateway, gin.H{
			"message": "画像のアップロードに失敗しました",
		})
		return
	}
	log.Printf("画像をアップロードできました！「%s」\n", imageName)

	// 投稿できたら、データベースに投稿したことを記録する
	// 挿入するデータを作成
	post := struct {
		ImagePath string `json:"image_path"`
		UserID    int    `json:"user_id"`
	}{
		ImagePath: imageName,
		UserID:    userID, // demo_userに固定してるよ
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "HTTPリクエストの作成に失敗しました",
		})
		return
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("apikey", API_KEY)
	req.Header.Set("Content-Type", "application/json")

	// HTTPリクエストを送信
	client = &http.Client{
		Timeout: 10 * time.Second, // タイムアウトを設定
	}
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

	// 👇️これが画像投稿した時に画像のURL返すJSONだよ
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
	url := API_URL + "/rest/v1/posts?select=image_path,posted_at,user_id" + "&user_id=eq." + strconv.Itoa(userID) + "&posted_at=gte." + startOfDay.UTC().Format(time.RFC3339) + "&posted_at=lt." + endOfDay.UTC().Format(time.RFC3339) + "&order=posted_at.desc&limit=1"

	// HTTPリクエストを作成
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "HTTPリクエストの作成に失敗しました",
		})
		return
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("apikey", API_KEY)

	// HTTPリクエストを送信
	client := &http.Client{
		Timeout: 10 * time.Second, // タイムアウトを設定
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "投稿の取得に失敗しました",
		})
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
		UserID    int    `json:"user_id"`
	}

	err = json.NewDecoder(resp.Body).Decode(&posts)
	if err != nil {
		log.Println(err)
		return
	}

	if len(posts) == 0 {
		log.Println("今日の投稿が見つかりませんでした。")
		// 👇️今日の投稿が無い時のプレーン画像のあれ
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "今日の投稿が見つかりませんでした。",
			"image_url": "/img/plane.png",
		})
		return
	}

	log.Printf("投稿を取得できました！:%+v\n", posts[0])

	// ストレージ上の画像のパスを返す
	imageURL := API_URL + "/storage/v1/object/public/" + strconv.Itoa(userID) + "/images/" + posts[0].ImagePath
	log.Printf("画像のURLを返します:%s\n", imageURL)

	// 👇️これが投稿した画像のJSONだよ
	c.JSON(http.StatusOK, gin.H{
		"image_url": imageURL, // 画像のURL
		// "posted_at": posts[0].PostedAt, // UTC時間での投稿日時返さなくても良いかなって
		// "user_id":   posts[0].UserID,   // ユーザID返さなくても良いかなって
	})
}
