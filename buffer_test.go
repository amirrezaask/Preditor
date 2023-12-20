package preditor

import (
	"fmt"
	"testing"
)

func TestSearch(t *testing.T) {
	matched := matchPatternCaseInsensitive([]byte("Hello World"), []byte("Hell"))

	_ = matched

	fmt.Println(matched)
}
