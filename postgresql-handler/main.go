package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/honeycombio/agentless-integrations-for-aws/common"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/honeycombio/honeytail/parsers/postgresql"
	"github.com/honeycombio/libhoney-go"
)

func main() {
	var err error
	if err = common.InitHoneycombFromEnvVars(); err != nil {
		logrus.WithError(err).
			Fatal("Unable to initialize libhoney with the supplied environment variables")
		return
	}
	defer libhoney.Close()
	common.AddUserAgentMetadata("rds", "postgresql")

	logLinePrefix := os.Getenv("LOG_LINE_PREFIX")
	parser := &postgresql.Parser{}
	parser.Init(&postgresql.Options{LogLinePrefix: logLinePrefix})

	dbh := common.NewPostgreSQLHandler(parser)
	dbh.Env = os.Getenv("ENVIRONMENT")

	if os.Getenv("SCRUB_QUERY") == "true" {
		dbh.ScrubQuery = true
	}

	lambda.Start(dbh.Handle)
}
