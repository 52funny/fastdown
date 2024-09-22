package fastdown

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Resume struct {
	ResumeLogs
	file *os.File
	lock sync.Mutex
	new  bool
}

type ResumeLogs struct {
	Concurrent int64   `json:"concurrent"`
	Ranges     []Range `json:"ranges"`
}

func NewResume(concurrent int64, path string, name string) (*Resume, error) {
	exist := exists(filepath.Join(path, name))
	var file *os.File
	switch exist {
	case true:
		f, err := os.Open(filepath.Join(path, name))
		if err != nil {
			return nil, err
		}
		file = f
	case false:
		f, err := os.Create(filepath.Join(path, name))
		if err != nil {
			return nil, err
		}
		file = f
	}
	new := !exist
	r := &Resume{
		file: file,
		lock: sync.Mutex{},
		new:  new,
	}
	if new {
		r.ResumeLogs = ResumeLogs{
			Concurrent: concurrent,
			Ranges:     make([]Range, concurrent),
		}
	} else {
		err := json.NewDecoder(file).Decode(&r.ResumeLogs)
		if err != nil {
			return nil, err
		}
	}
	if new || !new && json.NewDecoder(file).Decode(&r.ResumeLogs) != nil {
	}
	return r, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
