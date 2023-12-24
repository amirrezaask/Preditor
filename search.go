package preditor

import "unicode"

func matchPatternCaseInsensitive(data []byte, pattern []byte) [][]int {
	var matched [][]int
	var buf []byte
	start := -1
	for i, b := range data {

		if len(pattern) == len(buf) {
			matched = append(matched, []int{start, i - 1})
			buf = nil
			start = -1
		}
		idxToCheck := len(buf)
		if idxToCheck == 0 {
			start = i
		}
		if unicode.ToLower(rune(pattern[idxToCheck])) == unicode.ToLower(rune(b)) {
			buf = append(buf, b)
		} else {
			buf = nil
			start = -1
		}
	}

	return matched
}

func findNextMatch(data []byte, pattern []byte) []int {
	var buf []byte
	start := -1
	for i, b := range data {

		if len(pattern) == len(buf) {
			return []int{start, i - 1}
		}
		idxToCheck := len(buf)
		if idxToCheck == 0 {
			start = i
		}
		if unicode.ToLower(rune(pattern[idxToCheck])) == unicode.ToLower(rune(b)) {
			buf = append(buf, b)
		} else {
			buf = nil
			start = -1
		}
	}

	return nil
}

func matchPatternAsync(dst *[][]int, data []byte, pattern []byte) {
	go func() {
		*dst = matchPatternCaseInsensitive(data, pattern)
	}()
}
