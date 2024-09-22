package fastdown_test

import (
	"fmt"
	"testing"

	"github.com/52funny/fastdown"
	"github.com/stretchr/testify/assert"
)

func TestResume(t *testing.T) {
	r, err := fastdown.NewResume(8, "./", "abc.txt.resume")
	assert.Nil(t, err)
	fmt.Println(r)
}
