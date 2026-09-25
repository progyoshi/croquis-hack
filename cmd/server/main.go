package main

import (
	"fmt"
	"io"
	"net/http"
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

	// webフォルダをFrontendとして配信
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	fmt.Println("Server started: http://localhost:3000")

	//サーバ起動
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println(err)
	}

}
