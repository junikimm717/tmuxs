package main

func StringMap(arr []string) map[string]bool {
	res := make(map[string]bool)
	for _, elt := range arr {
		res[elt] = true
	}
	return res
}
