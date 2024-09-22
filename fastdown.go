package fastdown

import (
	"crypto/sha256"
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
	concurrent   int
	file         *os.File
	supportRange bool
	filePath     string
	fileName     string
	// default is os.TempDir()
	resumePath string
	// default is sha256(path + name)[:8]
	resumeName string
}

func NewDownloadWrapper(url string, concurrent int, path string, fileName string) (*DownloadWrapper, error) {
	hash := sha256.Sum256([]byte(filepath.Join(path, fileName)))
	return &DownloadWrapper{
		url:        url,
		length:     -1,
		concurrent: concurrent,
		file:       nil,
		filePath:   path,
		fileName:   fileName,
		resumePath: os.TempDir(),
		resumeName: fmt.Sprintf("%x", hash[:8]),
	}, nil
}

// Set the path of the resume file.
func (dw *DownloadWrapper) SetResumePath(path string) {
	dw.resumePath = path
}

// Download the file.
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

// Using io.Copy to download the file
func (dw *DownloadWrapper) normalDownload() error {
	f, err := os.Create(filepath.Join(dw.filePath, dw.fileName))
	if err != nil {
		return err
	}
	dw.file = f
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
	dw.file.Close()
	return nil
}

// Download the file with range, the concurrent is the number of the goroutines to download the file.
func (dw *DownloadWrapper) rangeDownload() error {
	exist := exists(filepath.Join(dw.resumePath, dw.resumeName))
	switch exist {
	case true:
		if !exists(filepath.Join(dw.filePath, dw.fileName)) {
			f, err := os.Create(filepath.Join(dw.filePath, dw.fileName))
			if err != nil {
				return err
			}
			dw.file = f
		} else {
			f, err := os.OpenFile(filepath.Join(dw.filePath, dw.fileName), os.O_RDWR, 0666)
			if err != nil {
				return err
			}
			dw.file = f
		}
	case false:
		f, err := os.Create(filepath.Join(dw.filePath, dw.fileName))
		if err != nil {
			return err
		}
		dw.file = f
	}

	// Close the file
	defer dw.file.Close()

	var resume *Resume
	switch exist {
	case false:
		ranges := make([]Range, dw.concurrent)
		chunkSize := (dw.length + int64(dw.concurrent) - 1) / int64(dw.concurrent)
		for i := 0; i < dw.concurrent; i++ {
			from := int64(i) * chunkSize
			to := from + chunkSize
			if from >= dw.length {
				break
			}
			// The last chunk, avoid the out of range
			if to > dw.length {
				to = dw.length
			}
			ranges[i] = NewRange(from, to)
		}
		re, err := NewResume(dw.resumePath, dw.resumeName, int(dw.concurrent), ranges)
		if err != nil {
			return err
		}
		resume = re
	case true:
		re, err := RecoverResume(dw.resumePath, dw.resumeName)
		if err != nil {
			return err
		}
		resume = re
	}

	// CLose the resume
	defer resume.Close()

	wait := sync.WaitGroup{}
	for i, r := range resume.Ranges {
		wait.Add(1)
		go dw.downloadChunk(i, r.From, r.To, resume, &wait)
	}
	wait.Wait()
	return nil
}

// Prepare the downloading, it will get the contentLength of the file and check if the server support range.
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("prepare status code not 200")
	}
	return nil
}

// Every goroutine download a chunk of the file.
func (dw *DownloadWrapper) downloadChunk(i int, from int64, to int64, re *Resume, wait *sync.WaitGroup) error {
	defer wait.Done()
	// defer fmt.Println("done", i, from, to)
	buf := make([]byte, 1024*8)
	l := from
	for l < to {
		req, err := http.NewRequest(http.MethodGet, dw.url, nil)
		if err != nil {
			continue
		}
		r := NewRange(l, to)
		req.Header.Add("Range", r.HeaderStr())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		for l < to {
			rn, err := resp.Body.Read(buf)
			if err != nil && err != io.EOF || rn == 0 {
				break
			}
			wn, err := dw.file.WriteAt(buf[:rn], l)
			if err != nil || wn != rn {
				break
			}
			l += int64(wn)
			re.Update(i, NewRange(l, to))
		}
		resp.Body.Close()
	}
	return nil
}
