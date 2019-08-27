package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

func TestGetReaderForKey(t *testing.T) {
	// Test basic gunzipping
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := &bytes.Buffer{}
		gw := gzip.NewWriter(b)
		gw.Write([]byte("gzipped payload"))
		gw.Close()

		w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "plain/text")
		w.WriteHeader(http.StatusOK)

		io.Copy(w, b)
	}))

	sess := session.Must(session.NewSession())

	svc := s3.New(sess, &aws.Config{
		Region:           aws.String(endpoints.UsWest2RegionID),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(server.URL),
	})

	reader, err := getReaderForKey(svc, "bucket", "key")
	assert.NoError(t, err)

	b := bytes.Buffer{}
	_, err = io.Copy(&b, reader)
	assert.NoError(t, err)
	reader.Close()

	assert.Equal(t, "gzipped payload", b.String())

	server.Close()
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := &bytes.Buffer{}
		b.Write([]byte("normal payload"))

		w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
		w.Header().Set("Content-Type", "plain/text")
		w.WriteHeader(http.StatusOK)

		io.Copy(w, b)
	}))

	svc = s3.New(sess, &aws.Config{
		Region:           aws.String(endpoints.UsWest2RegionID),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(server.URL),
	})

	reader, err = getReaderForKey(svc, "bucket", "key")
	assert.NoError(t, err)

	b = bytes.Buffer{}
	_, err = io.Copy(&b, reader)
	assert.NoError(t, err)
	reader.Close()

	assert.Equal(t, "normal payload", b.String())

	server.Close()
	// Now test a gzipped payload without a Content-Encoding header - This should fail until we set forceGunzip
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := &bytes.Buffer{}
		gw := gzip.NewWriter(b)
		gw.Write([]byte("gzipped payload"))
		gw.Close()

		w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
		w.Header().Set("Content-Type", "plain/text")
		w.WriteHeader(http.StatusOK)

		io.Copy(w, b)
	}))

	svc = s3.New(sess, &aws.Config{
		Region:           aws.String(endpoints.UsWest2RegionID),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(server.URL),
	})

	reader, err = getReaderForKey(svc, "bucket", "gzipped object no header")
	assert.NoError(t, err)

	b = bytes.Buffer{}
	_, err = io.Copy(&b, reader)
	assert.NoError(t, err)
	reader.Close()

	assert.NotEqual(t, "gzipped payload", b.String())

	// retry by forcing gunzip
	forceGunzip = true

	reader, err = getReaderForKey(svc, "bucket", "gzipped object no header")
	assert.NoError(t, err)

	b = bytes.Buffer{}
	_, err = io.Copy(&b, reader)
	assert.NoError(t, err)
	reader.Close()

	assert.Equal(t, "gzipped payload", b.String())

	server.Close()

	// Ensure fallback works when a non-gzipped object is picked up when forceGzipped is enabled
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := &bytes.Buffer{}
		b.Write([]byte("normal payload retry"))

		w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
		w.Header().Set("Content-Type", "plain/text")
		w.WriteHeader(http.StatusOK)

		io.Copy(w, b)
	}))

	svc = s3.New(sess, &aws.Config{
		Region:           aws.String(endpoints.UsWest2RegionID),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String(server.URL),
	})

	reader, err = getReaderForKey(svc, "bucket", "key")
	assert.NoError(t, err)

	b = bytes.Buffer{}
	_, err = io.Copy(&b, reader)
	assert.NoError(t, err)
	reader.Close()

	assert.Equal(t, "normal payload retry", b.String())

	server.Close()
}
