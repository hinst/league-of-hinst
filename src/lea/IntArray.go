package lea

func CheckIntArrayContains(a []int, value int) (result bool) {
	for _, item := range a {
		if item == value {
			result = true
			break
		}
	}
	return
}
