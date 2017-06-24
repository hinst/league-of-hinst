package lea

import "strconv"

func IntToStr(a int) string {
	return strconv.Itoa(a)
}

func RatioToStr(a, b int) string {
	var ratio = float32(a) / float32(b)
	return IntToStr(int(ratio * 100))
}
