package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/honeycombio/honeytail/parsers/postgresql"
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/transmission"
)

type tWriter struct {
	T *testing.T
}

func (w tWriter) Write(b []byte) (int, error) {
	w.T.Logf("%s", b)
	return len(b), nil
}

func TestParser(t *testing.T) {
	lg := common.LoggerFromEnv()
	lg.SetOutput(tWriter{T: t})

	parser := &postgresql.Parser{}
	err := parser.Init(&postgresql.Options{
		LogLinePrefix: "%t:%r:%u@%d:[%p]:",
	})
	require.NoError(t, err)

	mock := transmission.MockSender{}
	c, err := libhoney.NewClient(libhoney.ClientConfig{
		APIKey:       "xyz",
		Dataset:      "abc",
		SampleRate:   1,
		APIHost:      "http://localhost:9999",
		Transmission: &mock,
		Logger:       lg.WithField("source", "libhoney"),
	})
	require.NoError(t, err)

	dbh := common.NewPostgreSQLHandler(parser)
	dbh.HoneyClient = c
	dbh.Logger = lg

	f, err := os.Open("../fixtures/pg-cloudwatch.json")
	require.NoError(t, err)

	var evts events.CloudwatchLogsEvent
	err = json.NewDecoder(f).Decode(&evts)
	require.NoError(t, err)

	res, err := dbh.Handle(evts)
	assert.NoError(t, err)
	assert.True(t, res.Ok)

	sent := mock.Events()
	if assert.Len(t, sent, 2) {
		q := "select id, color from colors where id = ?;"
		assert.Equal(t, q, sent[0].Data["normalized_query"])
		assert.Equal(t, q, sent[1].Data["normalized_query"])
	}
}
