package group

import "strings"

const Default = "default"

func Normalize(value string) string {
	if strings.TrimSpace(value) == "" {
		return Default
	}
	return value
}

