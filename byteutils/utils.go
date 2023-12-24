package byteutils

import "unicode"

func PreviousWordInBuffer(bs []byte, idx int) int {

	var firstNonWord int
	for i := idx; i >= 0; i-- {
		if len(bs) > i && !unicode.IsLetter(rune(bs[i])) {
			firstNonWord = i
			break
		}
	}
	if len(bs) <= idx+1-firstNonWord {
		return -1
	}
	var firstWordEnd int
	for i := idx - firstNonWord; i >= 0; i-- {
		if unicode.IsLetter(rune(bs[i])) {
			firstWordEnd = i
			break
		}
	}

	if firstWordEnd == 0 {
		firstWordEnd = 1
	}

	idx = idx - (firstWordEnd + 1)
	if idx < 0 {
		idx = 0
	}
	return idx
}

func NextWordInBuffer(bs []byte, idx int) int {
	var firstNonWord int
	if idx+1 >= len(bs) {
		return -1
	}
	for idx, b := range bs[idx+1:] {
		if !unicode.IsLetter(rune(b)) {
			firstNonWord = idx
			break
		}
	}
	if len(bs) <= idx+1+firstNonWord {
		idx = len(bs)
		return -1
	}
	var jump int
	for i, b := range bs[idx+1+firstNonWord:] {
		if unicode.IsLetter(rune(b)) {
			jump = 1 + firstNonWord + i
			break
		}
	}
	idx = idx + jump

	if idx > len(bs) || jump == 1 {
		idx = len(bs)
	}
	return idx
}
