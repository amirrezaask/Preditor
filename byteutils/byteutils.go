package byteutils

import (
	"unicode"
)

const nonLetter = "~!@#$%^&*()_+{}[];:'\"\\\n\r <>,.?/"

func SeekNextNonLetter(bs []byte, idx int) int {
	for i := idx + 1; i < len(bs); i++ {
		if i == len(bs) {
			continue
		}
		if !unicode.IsLetter(rune(bs[i])) {
			return i
		}
	}

	return -1
}

func SeekPreviousNonLetter(bs []byte, idx int) int {
	for i := idx - 1; i >= 0; i-- {
		if i == len(bs) {
			continue
		}
		if !unicode.IsLetter(rune(bs[i])) {
			return i
		}
	}
	return -1
}

func SeekPreviousLetter(bs []byte, idx int) int {
	for i := idx - 1; i < len(bs); i++ {
		if i == len(bs) {
			continue
		}
		if unicode.IsLetter(rune(bs[i])) {
			return i
		}
	}
	return -1
}
func SeekNextLetter(bs []byte, idx int) int {
	for i := idx + 1; i >= 0; i-- {
		if i == len(bs) {
			continue
		}
		if unicode.IsLetter(rune(bs[i])) {
			return i
		}
	}
	return -1
}

func PreviousWordInBuffer(bs []byte, idx int) int {
	lastNonLetter := SeekPreviousNonLetter(bs, idx)
	lastLetterIndex := SeekPreviousLetter(bs, lastNonLetter)
	return lastLetterIndex
}

func NextWordInBuffer(bs []byte, idx int) int {
	nextNonLetter := SeekNextNonLetter(bs, idx)
	nextLetterIndex := SeekNextLetter(bs, nextNonLetter)
	return nextLetterIndex
}
