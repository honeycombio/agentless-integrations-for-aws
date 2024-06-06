package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RequestRedactSuite struct {
	suite.Suite
}

func (suite *RequestRedactSuite) TestRedact() {
	sampleRequest := "GET /some/path?key1=value1&key2=value2&keyN=valueN HTTP/1.1"
	redactPattern := regexp.MustCompile(`key\d=(value)\d`)

	redactedRequest := redactRequest(sampleRequest, redactPattern)
	suite.Equal("GET /some/path?key1=xxxxx1&key2=xxxxx2&keyN=valueN HTTP/1.1", redactedRequest)
}

func TestRequestRedaction(t *testing.T) {
	suite.Run(t, new(RequestRedactSuite))
}
