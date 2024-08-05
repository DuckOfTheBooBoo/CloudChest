package utils

import (
	"strings"
)

func GetFileExtension(filename string) string {
	index := strings.LastIndex(filename, ".")
	if index == -1 {
			return ""
	}
	return filename[index+1:]
}
