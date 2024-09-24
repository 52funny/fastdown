package fastdown_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/52funny/fastdown"
)

func TestDownload(t *testing.T) {
	url := "https://dl-a10b-0625.mypikpak.com/download/?fid=tM1wTRN2lj6zBlJlVqUr2aKJwuisVZulya_VcsWmBzfXJZdtoAqKltYmUT8GG7Oq-KmCCCSW2Z3vamrhyrEIm2DoZzT-vUu_tJ2oJhKlEDU=&from=5&verno=3&prod=pikpak&expire=1727232880&g=E1E351E5B7161BBF37E37E9D44D99C28F32E4F75&ui=Yu4fyocWcVXMccgO&t=0&ms=5040000&th=5040000&f=1873280858&alt=0&us=0&hspu=&po=0&fileid=VO6pUjWQJdcCass2BMamWYYro1&userid=Yu4fyocWcVXMccgO&pr=XQPkPvr9WWiIuMvELmrVerTUTLG_pIgrFR4yWrc_dqAxC6Uk7qOAnX-JYbrAFNvTG_7uY-kHtK7NdWKFMi0hwoO7Fm1kb0LJjkpfO_3a776Q-7EaX1qauwiQTEAp_4DTozk-JW8bwTLCfoxkCOJ-4LnprOLGCSK-gQyBRx3amBWC5HwkUqpeEPYhZCc0WOu8_IrBgTImqeNA-JKtKRCi60nWKZBDqVwgmJY4QvDU9eFnzNhAEg1ugFkTULfSKrm_3nE5ATMHc1g-OxurQb5rE8xnLGslcIoDtaMewHsFCH3vSXcS4rrwRhjEjF3NQemDkzIktyRyIzjErmvq1DoAQcVm6pm5ysX11RlCcutI3930XMN9D1bOcwc7I7JqTEUFpm9YJS4tTeS28LJhnRCKWKeedaQjp70z_N81Pt4EP5A=&sign=27879A3FCD97015B35247750391D6384"
	len := getContentLen(url)
	fmt.Println(len)
	dw := fastdown.NewDownloadWrapper(url, 1, len, "./", "test.txt")
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
