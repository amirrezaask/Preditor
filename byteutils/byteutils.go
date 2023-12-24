package byteutils

import (
	"unicode"
)

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
	var sawWord bool
	var sawWhitespaces bool
	for i := idx - 1; i >= 0; i-- {
		if sawWord {
			return i + 1
		}
		if i == 0 {
			return i
		}

		if !unicode.IsLetter(rune(bs[i])) {
			sawWhitespaces = true
		} else {
			if sawWhitespaces {
				sawWord = true
				if sawWord && sawWhitespaces {
					return i
				}
			}
		}
	}

	return -1

}
func NextWordInBuffer(bs []byte, idx int) int {
	var sawWord bool
	var sawWhitespaces bool
	for i := idx + 1; i < len(bs); i++ {
		if sawWord {
			return i + 1
		}
		if !unicode.IsLetter(rune(bs[i])) {
			sawWhitespaces = true
		} else {
			if sawWhitespaces {
				sawWord = true
				if sawWord && sawWhitespaces {
					return i
				}
			}
		}
	}

	return -1
}
