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

func FindMatchingClosedForward(data []byte, idx int) int {
	in := data[idx]
	var matching byte
	if in == '[' {
		matching = ']'
	} else if in == '(' {
		matching = ')'
	} else if in == '{' {
		matching = '}'
	} else {
		return -1
	}

	getCharDelta := func(c byte) int {
		if c == in {
			return 1
		} else if c == matching {
			return -1
		} else {
			return 0
		}
	}

	state := 1
	for i := idx + 1; i < len(data); i++ {
		c := data[i]
		state += getCharDelta(c)
		if state == 0 {
			return i
		}
	}

	return -1
}
func FindMatchingOpenBackward(data []byte, idx int) int {
	in := data[idx]
	var matching byte
	if in == ')' {
		matching = '('
	} else if in == ']' {
		matching = '['
	} else if in == '}' {
		matching = '{'
	} else {
		return -1
	}

	getCharDelta := func(c byte) int {
		if c == in {
			return -1
		} else if c == matching {
			return 1
		} else {
			return 0
		}
	}

	state := -1
	for i := idx - 1; i >= 0; i-- {
		c := data[i]
		state += getCharDelta(c)
		if state == 0 {
			return i
		}
	}

	return -1
}

func FindMatching(data []byte, idx int) int {
	if len(data) == 0 {
		return -1
	}
	switch data[idx] {
	case '{', '[', '(':
		return FindMatchingClosedForward(data, idx)
	case '}', ']', ')':
		return FindMatchingOpenBackward(data, idx)
	default:
		return -1
	}
}
