package main

import (
	"regexp"
	"strings"
)

func redactRequest(request string, redactPattern *regexp.Regexp) string {
	matches := redactPattern.FindAllSubmatchIndex([]byte(request), -1)
	for _, match := range matches {
		for i := 1; i < len(match)/2; i++ {
			start := match[i*2]
			end := match[i*2+1]
			request = request[:start] + strings.Repeat("x", end-start) + request[end:]
		}
	}
	return request
}
