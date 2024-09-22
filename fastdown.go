package fastdown

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type DownloadWrapper struct {
	url          string
	length       int64
	concurrent   int64
	file         *os.File
	supportRange bool
}

func NewDownloadWrapper(url string, concurrent int64, path string, name string) (*DownloadWrapper, error) {
	file, err := os.Create(filepath.Join(path, name))
	if err != nil {
		return nil, err
	}
	return &DownloadWrapper{
		url:        url,
		length:     -1,
		concurrent: concurrent,
		file:       file,
	}, nil
}

func (dw *DownloadWrapper) Download() error {
	// Prepare the downloading
	err := dw.perpare()
	if err != nil {
		return err
	}
	switch dw.supportRange {
	case true:

		return dw.rangeDownload()
	case false:
		return dw.normalDownload()
	}
	return nil
}

func (dw *DownloadWrapper) normalDownload() error {
	resp, err := http.Get(dw.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(dw.file, resp.Body)
	if err != nil {
		dw.file.Close()
		os.Remove(dw.file.Name())
		return err
	}
	return nil
}

func (dw *DownloadWrapper) rangeDownload() error {
	// Close the file
	defer dw.file.Close()

	chunkSize := (dw.length + dw.concurrent - 1) / dw.concurrent
	fmt.Println("chunkSize", chunkSize)
	wait := sync.WaitGroup{}
	for i := int64(0); i < dw.concurrent; i++ {
		from := i * chunkSize
		to := from + chunkSize
		if from >= dw.length {
			break
		}
		if to > dw.length {
			to = dw.length
		}
		fmt.Println(from, to)
		wait.Add(1)
		go dw.downloadChunk(from, to, &wait)
	}
	wait.Wait()
	return nil
}

func (dw *DownloadWrapper) perpare() error {
	resp, err := http.Get(dw.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dw.length = resp.ContentLength
	ranges := resp.Header.Get("Accept-Ranges")
	if ranges == "bytes" {
		dw.supportRange = true
	}
	fmt.Println("contentLength", dw.length)
	return nil
}

func (dw *DownloadWrapper) downloadChunk(from int64, to int64, wait *sync.WaitGroup) error {
	defer wait.Done()
	defer fmt.Println("done", from, to)
	buf := make([]byte, 1024*8)
	// writren := 0
	l := from
	for l < to {
		req, err := http.NewRequest(http.MethodGet, dw.url, nil)
		if err != nil {
			continue
		}
		r := NewRange(l, to)
		req.Header.Add("Range", r.String())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		for l < to {
			rn, err := resp.Body.Read(buf)
			if err != nil && err != io.EOF {
				break
			}
			if rn == 0 {
				break
			}
			wn, err := dw.file.WriteAt(buf[:rn], l)

			l += int64(wn)
			if err != nil {
				break
			}
		}
		resp.Body.Close()
	}
	return nil
}
