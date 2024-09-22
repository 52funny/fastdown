package fastdown_test

import (
	"testing"

	"github.com/52funny/fastdown"
	"github.com/stretchr/testify/assert"
)

func TestResume(t *testing.T) {
	ranges := []fastdown.Range{
		{0, 10},
		{10, 20},
		{20, 30},
		{30, 40},
		{50, 60},
		{60, 70},
		{70, 80},
		{80, 90},
	}
	r, err := fastdown.NewResume("./", "abc.txt.resume", 8, ranges)
	assert.Nil(t, err)
	defer r.Close()
	assert.Equal(t, 8, r.Concurrent)

	r.Update(0, fastdown.Range{0, 5})
	assert.Equal(t, fastdown.Range{0, 5}, r.Ranges[0])
}
