package common

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/honeytail/parsers/mysql"
	"github.com/honeycombio/honeytail/parsers/postgresql"
	libhoney "github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/transmission"
)

type tWriter struct {
	T *testing.T
}

func (w tWriter) Write(b []byte) (int, error) {
	w.T.Logf("%s", b)
	return len(b), nil
}

func mockedHoneyClient(t *testing.T, lg logrus.FieldLogger) (*transmission.MockSender, *libhoney.Client) {
	mock := &transmission.MockSender{}
	client, err := libhoney.NewClient(libhoney.ClientConfig{
		APIKey:       "xyz",
		Dataset:      "abc",
		SampleRate:   1,
		APIHost:      "http://localhost:9999",
		Transmission: mock,
		Logger:       lg.WithField("source", "libhoney"),
	})
	require.NoError(t, err)
	return mock, client
}

type testcase struct {
	Name           string
	HandlerBuilder func() (*DBHandler, error)
	FixtureFile    string
	ExpectedQuery  string
}

func TestHandle_PostgresParser(t *testing.T) {
	lg := LoggerFromEnv()
	lg.SetOutput(tWriter{T: t})

	for _, tc := range []testcase{
		{
			Name:          "with Postgres parser",
			FixtureFile:   "pg-cloudwatch.json",
			ExpectedQuery: "select id, color from colors where id = ?;",
			HandlerBuilder: func() (*DBHandler, error) {
				parser := &postgresql.Parser{}
				err := parser.Init(&postgresql.Options{
					LogLinePrefix: "%t:%r:%u@%d:[%p]:",
				})
				return NewPostgreSQLHandler(parser), err
			},
		}, {
			Name:          "with Mysql parser",
			FixtureFile:   "mysql-cloudwatch.json",
			ExpectedQuery: "select * from colors where id = ?",
			HandlerBuilder: func() (*DBHandler, error) {
				parser := &mysql.Parser{}
				err := parser.Init(&mysql.Options{})
				return NewMySQLHandler(parser), err
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			mock, client := mockedHoneyClient(t, lg)
			dbh, err := tc.HandlerBuilder()
			require.NoError(t, err)
			dbh.HoneyClient = client
			dbh.Logger = lg

			f, err := os.Open("../fixtures/" + tc.FixtureFile)
			require.NoError(t, err)

			var evts events.CloudwatchLogsEvent
			err = json.NewDecoder(f).Decode(&evts)
			require.NoError(t, err)

			res, err := dbh.Handle(evts)
			assert.NoError(t, err)
			assert.True(t, res.Ok)

			sent := mock.Events()
			if assert.Len(t, sent, 2) {
				assert.Equal(t, tc.ExpectedQuery, sent[0].Data["normalized_query"])
				assert.Equal(t, tc.ExpectedQuery, sent[1].Data["normalized_query"])
			}
		})
	}
}
