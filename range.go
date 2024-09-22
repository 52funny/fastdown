package fastdown

import "fmt"

// The scope of the Range is [from, to),
// it's allow from == to, which means the range is empty
type Range struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

// Display the Range in the form of [from, to)
func (r Range) String() string {
	return fmt.Sprintf("[%d, %d)", r.From, r.To)
}

// This function is used to generate the Range header string
func (r Range) HeaderStr() string {
	return fmt.Sprintf("bytes=%d-%d", r.From, r.To-1)
}

// Create a new Range
func NewRange(from int64, to int64) Range {
	if from < 0 || from > to {
		panic("invalid range " + fmt.Sprintf("from: %d, to: %d", from, to))
	}
	return Range{
		From: from,
		To:   to,
	}
}
