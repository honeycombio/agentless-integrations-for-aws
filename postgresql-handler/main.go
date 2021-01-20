package main

import (
	"os"

	"github.com/honeycombio/agentless-integrations-for-aws/common"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/honeycombio/honeytail/parsers/postgresql"
	"github.com/honeycombio/libhoney-go"
)

func main() {
	logger := common.LoggerFromEnv()

	if err := common.InitHoneycombFromEnvVars(); err != nil {
		logger.WithError(err).
			Fatal("Unable to initialize libhoney with the supplied environment variables")
	}
	defer libhoney.Close()
	common.AddUserAgentMetadata("rds", "postgresql")

	parser := &postgresql.Parser{}
	parser.Init(&postgresql.Options{
		LogLinePrefix: os.Getenv("LOG_LINE_PREFIX"),
	})

	dbh := common.NewPostgreSQLHandler(parser)
	dbh.Env = os.Getenv("ENVIRONMENT")
	dbh.Logger = logger

	if os.Getenv("SCRUB_QUERY") == "true" {
		dbh.ScrubQuery = true
	}

	lambda.Start(dbh.Handle)
}
