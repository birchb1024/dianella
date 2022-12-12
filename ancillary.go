package dianella

// stringTruncate - return the first <length> runes of the string
func stringTruncate(text string, length uint) string {
	r := []rune(text)
	trunc := r[:intMin(len(r), int(length))]
	result := string(trunc)
	if len(result) < len(text) {
		return result + "..."
	}
	return result
}

// intMin - return the smallest integer
func intMin(x, y int) int {
	if x > y {
		return y
	}
	return x
}
