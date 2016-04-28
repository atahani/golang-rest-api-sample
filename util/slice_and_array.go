package util

//the helper functions for slice and arrays
func IsStringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
