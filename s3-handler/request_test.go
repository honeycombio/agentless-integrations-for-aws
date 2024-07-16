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

type RequestUrlParseSuite struct {
	suite.Suite
}

func (suite *RequestUrlParseSuite) TestParseUrl() {
	sampleRequest := "GET https://sample.host.tld:443/v1/products/foobar/events?param1=xxxx234 HTTP/2.0"
	httpMeta, err := parseRequestHttpMeta(sampleRequest)
	suite.Nil(err)
	suite.Equal("GET", httpMeta["http.request.method"])
	suite.Equal("https", httpMeta["url.scheme"])
	suite.Equal("sample.host.tld", httpMeta["server.address"])
	suite.Equal("443", httpMeta["server.port"])
	suite.Equal("/v1/products/foobar/events", httpMeta["url.path"])
	suite.Equal("param1=xxxx234", httpMeta["url.query"])
	suite.Equal("2.0", httpMeta["network.protocol.version"])
}

func TestRequestUrlParsing(t *testing.T) {
	suite.Run(t, new(RequestUrlParseSuite))
}
