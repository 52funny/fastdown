package fastdown_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/52funny/fastdown"
)

func TestDownload(t *testing.T) {
	url := "https://speedtest1.online.sh.cn:8080/download?size=1024&r=0.33838968640999534"
	len := getContentLen(url)
	fmt.Println(len)
	dw := fastdown.NewDownloadWrapper(url, 1, 1024, "./", "test.txt")
	err := dw.Download()
	if err != nil {
		t.Error(err)
	}
}

func getContentLen(url string) int64 {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	return resp.ContentLength
}
