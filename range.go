package fastdown

import "fmt"

// The scope of the Range is [from, to)
type Range struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

func (r Range) String() string {
	return fmt.Sprintf("bytes=%d-%d", r.From, r.To-1)
}

func NewRange(from int64, to int64) Range {
	if from < 0 || from > to {
		panic("invalid range " + fmt.Sprintf("from: %d, to: %d", from, to))
	}
	return Range{
		From: from,
		To:   to,
	}
}
