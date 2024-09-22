package fastdown

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
)

type Resume struct {
	ResumeLogs
	file *os.File
	new  bool
}

type ResumeLogs struct {
	Concurrent int     `json:"concurrent"`
	Ranges     []Range `json:"ranges"`
}

func NewResume(path string, name string, concurrent int, ranges []Range) (*Resume, error) {
	if concurrent != len(ranges) {
		panic("concurrent should be equal to the number of ranges")
	}
	var file *os.File
	f, err := os.Create(filepath.Join(path, name))
	if err != nil {
		return nil, err
	}
	file = f
	r := &Resume{
		file: file,
		new:  true,
		ResumeLogs: ResumeLogs{
			Concurrent: concurrent,
			Ranges:     ranges,
		},
	}
	buf := make([]byte, 4+concurrent*16)
	binary.LittleEndian.PutUint32(buf, uint32(r.Concurrent))
	for i := 0; i < concurrent; i++ {
		binary.LittleEndian.PutUint64(buf[4+i*16:], uint64(ranges[i].From))
		binary.LittleEndian.PutUint64(buf[4+i*16+8:], uint64(ranges[i].To))
	}
	buffer := bytes.NewBuffer(buf)
	io.Copy(file, buffer)
	return r, nil
}

func RecoverResume(path string, name string) (*Resume, error) {
	f, err := os.Open(filepath.Join(path, name))
	if err != nil {
		return nil, err
	}
	r := &Resume{
		file: f,
		new:  false,
	}

	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	concurrent := binary.LittleEndian.Uint32(buf[:4])
	r.Concurrent = int(concurrent)
	buf = buf[4:]

	ranges := make([]Range, r.Concurrent)
	for i := 0; i < r.Concurrent; i++ {
		ranges[i].From = int64(binary.LittleEndian.Uint64(buf[0:8]))
		ranges[i].To = int64(binary.LittleEndian.Uint64(buf[8:16]))
		buf = buf[16:]
	}
	r.Ranges = ranges
	// fmt.Println("resume ranges", ranges)
	return r, nil
}

func (resume *Resume) Update(i int, r Range) error {
	resume.Ranges[i] = r
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[:8], uint64(r.From))
	binary.LittleEndian.PutUint64(buf[8:], uint64(r.To))
	wn, err := resume.file.WriteAt(buf, 4+int64(i)*16)
	if err != nil {
		return err
	}
	if wn != 16 {
		return errors.New("update resume bin file wn not equals 16")
	}
	return nil
}

func (resume *Resume) Close() error {
	err := resume.file.Close()
	if err != nil {
		return err
	}
	// Remove resume file if all ranges are downloaded
	return os.Remove(resume.file.Name())
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
