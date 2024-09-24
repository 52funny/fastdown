package fastdown

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type DownloadWrapper struct {
	url           string
	contentLength int64
	concurrent    int
	file          *os.File
	resume        *Resume
	supportRange  bool
	filePath      string
	fileName      string
	// default is os.TempDir()
	resumePath string
	// default is sha256(path + name)[:8]
	resumeName string
}

var DefaultClient = http.Client{}
var defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"

func NewDownloadWrapper(url string, concurrent int, contentLength int64, path string, fileName string) *DownloadWrapper {
	hash := sha256.Sum256([]byte(filepath.Join(path, fileName)))
	return &DownloadWrapper{
		url:           url,
		contentLength: contentLength,
		concurrent:    concurrent,
		file:          nil,
		filePath:      path,
		fileName:      fileName,
		resumePath:    os.TempDir(),
		resumeName:    fmt.Sprintf("%x", hash[:8]),
	}
}

func (dw *DownloadWrapper) prepare() error {
	existResume := exists(filepath.Join(dw.resumePath, dw.resumeName))
	existFile := exists(filepath.Join(dw.filePath, dw.fileName))

	// The download strategy:
	// +-------------+-----------+------------------+
	// | existResume | existFile | op               |
	// +=============+===========+==================+
	// | 1           | 1         | Resume download  |
	// | 1           | 0         | Restart download |
	// | 0           | 1         | Restart download |
	// | 0           | 0         | Restart download |
	// +-------------+-----------+------------------+

	if existResume && existFile {
		// Recover the download
		re, err := RecoverResume(dw.resumePath, dw.resumeName)
		if err != nil {
			return err
		}
		if re.Concurrent != dw.concurrent {
			re.Close()
			re.Remove()
			return dw.startNewDownload()
		}
		dw.resume = re

		f, err := os.OpenFile(filepath.Join(dw.filePath, dw.fileName), os.O_RDWR, 0666)
		if err != nil {
			return err
		}
		dw.file = f
	} else {
		return dw.startNewDownload()
	}
	return nil
}

func (dw *DownloadWrapper) startNewDownload() error {
	f, err := os.Create(filepath.Join(dw.filePath, dw.fileName))
	if err != nil {
		return err
	}
	dw.file = f

	// Create the new ranges and restart the download
	ranges := make([]Range, dw.concurrent)
	chunkSize := (dw.contentLength + int64(dw.concurrent) - 1) / int64(dw.concurrent)
	for i := 0; i < dw.concurrent; i++ {
		from := int64(i) * chunkSize
		to := from + chunkSize
		if from >= dw.contentLength {
			break
		}
		// The last chunk, avoid the out of range
		if to > dw.contentLength {
			to = dw.contentLength
		}
		ranges[i] = NewRange(from, to)
	}
	re, err := NewResume(dw.resumePath, dw.resumeName, int(dw.concurrent), ranges)
	if err != nil {
		return err
	}
	dw.resume = re
	return nil
}

// Set the path of the resume file.
func (dw *DownloadWrapper) SetResumePath(path string) {
	dw.resumePath = path
}

// Download the file.
func (dw *DownloadWrapper) Download() error {
	return dw.rangeDownload()
}

// Download the file with range, the concurrent is the number of the goroutines to download the file.
func (dw *DownloadWrapper) rangeDownload() error {
	err := dw.prepare()
	if err != nil {
		return fmt.Errorf("prepare failed: %v", err)
	}

	// Close the file
	defer dw.file.Close()

	// CLose the resume
	defer dw.resume.Close()

	wait := sync.WaitGroup{}
	var counter int32
	for i, r := range dw.resume.Ranges {
		wait.Add(1)
		go dw.downloadChunk(i, r.From, r.To, &wait, &counter)
	}
	wait.Wait()
	if counter == int32(dw.resume.Concurrent) {
		dw.resume.Remove()
	} else {
		return fmt.Errorf("download not completed")
	}
	return nil
}

// Every goroutine download a chunk of the file.
func (dw *DownloadWrapper) downloadChunk(i int, from int64, to int64, wait *sync.WaitGroup, counter *int32) error {
	defer wait.Done()

	if from == to {
		atomic.AddInt32(counter, 1)
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, dw.url, nil)
	if err != nil {
		return err
	}
	r := NewRange(from, to)
	req.Header.Add("Range", r.HeaderStr())
	req.Header.Add("User-Agent", defaultUserAgent)
	resp, err := DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download chunk %d failed: %v", i, err)
	}
	defer resp.Body.Close()
	writer := partialWriter{
		Writer:       io.NewOffsetWriter(dw.file, from),
		PartialIndex: i,
		Resume:       dw.resume,
		Pos:          from,
		Last:         to,
	}
	wn, err := io.Copy(&writer, resp.Body)
	if err != nil {
		return err
	}

	if wn == to-from {
		atomic.AddInt32(counter, 1)
	}
	return nil
}

type partialWriter struct {
	Writer       io.Writer
	PartialIndex int
	Resume       *Resume
	Pos          int64
	Last         int64
}

func (w *partialWriter) Write(p []byte) (n int, err error) {
	if w.Pos >= w.Last {
		return 0, io.EOF
	}
	if int64(len(p)) > w.Last-w.Pos {
		p = p[:w.Last-w.Pos]
	}
	n, err = w.Writer.Write(p)
	w.Pos += int64(n)
	w.Resume.Update(w.PartialIndex, NewRange(w.Pos, w.Last))
	return
}
