package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 画像を受け取るテスト
func testHandler(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("test")
	if err == nil {
		data, err := io.ReadAll(file)
		if err == nil {
			fmt.Fprintln(w, string(data))
		}
	}
}

func main() {
	// http.HandleFunc("/", testHandler)

	router := gin.Default() // Ginルーター初期化

	// webフォルダをFrontendとして配信
	router.Static("/", "./web")

	// 標準出力にメッセージ表示
	fmt.Println("Server started: http://localhost:3000")

	//サーバ起動
	router.Run(":3000")

}
