package go_utils

func StringInSlice(s string, list []string) bool {
	for _, value := range list {
		if value == s {
			return true
		}
	}
	return false
}
