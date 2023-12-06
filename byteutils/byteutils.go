package byteutils

import (
	"bytes"
)

func SeekNextWhitespace(bs []byte, i int) int {
	idx := bytes.IndexAny(bytes.TrimLeft(bs[i:], " \n\r"), " \n\r")
	return i + idx
}

func SeekPreviousWhitespace(bs []byte, i int) int {
	idx := bytes.LastIndexAny(bs[:i], " \n\r")
	return idx
}

func PreviousWordInBuffer(bs []byte, idx int) int {
	lastWhitespaceIndex := SeekPreviousWhitespace(bs, idx)
	return lastWhitespaceIndex - 1
}

func NextWordInBuffer(bs []byte, idx int) int {
	firstWhitespaceIndex := SeekNextWhitespace(bs, idx)
	return firstWhitespaceIndex + 1
}
