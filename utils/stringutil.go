package utils

import (
	"strings"
)

func SplitAndRemoveSpace(s, sep string) []string {
	array := []string{}
	for _, el := range strings.Split(s, sep) {
		//array[i] = strings.Replace(el, " ", "", -1)
		array = append(array, strings.Replace(el, " ", "", -1))
	}
	return array
}
