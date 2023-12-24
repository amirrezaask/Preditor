package byteutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindMatchingClosedForward(t *testing.T) {
	assert.Equal(t, 5, FindMatchingClosedForward([]byte(`({[}])`), 0))
	assert.Equal(t, 3, FindMatchingClosedForward([]byte(`({[}])`), 1))
	assert.Equal(t, 4, FindMatchingClosedForward([]byte(`({[}])`), 2))
}

func TestFindMatchingOpenBackward(t *testing.T) {
	assert.Equal(t, 0, FindMatchingOpenBackward([]byte(`({[}])`), 5))
	assert.Equal(t, 1, FindMatchingOpenBackward([]byte(`({[}])`), 3))
	assert.Equal(t, 2, FindMatchingOpenBackward([]byte(`({[}])`), 4))
}
