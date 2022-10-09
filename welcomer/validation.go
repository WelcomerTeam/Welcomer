package welcomer

import "strconv"

const (
	hexBase      = 16
	int64Base    = 10
	int64BitSize = 64
)

func isValidUnsplashID(str string) bool {
	return unsplashRegex.MatchString(str)
}

func isValidColour(str string) bool {
	_, err := ParseColour(str, "")

	return err == nil
}

func isValidHex(str string, allowAlpha bool) bool {
	if len(str) != 6 && (!allowAlpha || len(str) != 8) {
		return false
	}

	_, err := strconv.ParseUint(str, hexBase, int64BitSize)

	return err == nil
}
