package byteutils

import (
	"unicode"
)

const nonLetter = "~!@#$%^&*()_+{}[];:'\"\\\n\r <>,.?/"

func SeekNextNonLetter(bs []byte, idx int) int {
	for i := idx; i < len(bs); i++ {
		if !unicode.IsLetter(rune(bs[i])) {
			return i
		}
	}

	return -1
}

func SeekPreviousNonLetter(bs []byte, idx int) int {
	for i := idx; i >= 0; i-- {
		if !unicode.IsLetter(rune(bs[i])) {
			return i
		}
	}
	return -1
}

func PreviousWordInBuffer(bs []byte, idx int) int {
	lastWhitespaceIndex := SeekPreviousNonLetter(bs, idx)
	return lastWhitespaceIndex - 1
}

func NextWordInBuffer(bs []byte, idx int) int {
	firstWhitespaceIndex := SeekNextNonLetter(bs, idx)
	return firstWhitespaceIndex + 1
}
