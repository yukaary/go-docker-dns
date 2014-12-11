package utils

import (
	"regexp"
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

func SplitScaledHostname(hostname string) (string, string) {
	pattern := regexp.MustCompile("(.*)_([0-9]+)$")
	group := pattern.FindStringSubmatch(hostname)
	if len(group) < 2 {
		return hostname, ""
	}
	return group[1], group[2]
}
