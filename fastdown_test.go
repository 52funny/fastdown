package fastdown_test

import (
	"testing"

	"github.com/52funny/fastdown"
)

func TestDownload(t *testing.T) {
	url := "https://speedtest1.online.sh.cn:8080/download?size=1024&r=0.33838968640999534"
	dw, err := fastdown.NewDownloadWrapper(url, 8, "./", "test.txt")
	if err != nil {
		t.Error(err)
	}
	dw.Download()
}

