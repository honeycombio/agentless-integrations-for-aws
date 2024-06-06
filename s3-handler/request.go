package main

import (
	"net/url"
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

var HTTP_VERSION_PATTERN = regexp.MustCompile(`HTTP/(\d+\.?\d*)`)

func parseRequestHttpMeta(request string) (map[string]string, error) {
	httpMeta := make(map[string]string)
	parts := strings.Split(request, " ")
	httpMeta["http.request.method"] = parts[0]
	if httpVersion := HTTP_VERSION_PATTERN.FindStringSubmatch(parts[2]); httpVersion != nil && len(httpVersion) > 1 {
		httpMeta["network.protocol.version"] = httpVersion[1]
	}

	parsedUrl, err := url.Parse(parts[1])
	if err != nil {
		return httpMeta, err
	}

	httpMeta["url.scheme"] = parsedUrl.Scheme
	httpMeta["server.address"] = parsedUrl.Hostname()
	httpMeta["server.port"] = parsedUrl.Port()
	httpMeta["url.path"] = parsedUrl.Path
	httpMeta["url.query"] = parsedUrl.RawQuery

	return httpMeta, nil
}
