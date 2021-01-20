package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/honeycombio/agentless-integrations-for-aws/common"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/honeycombio/honeytail/parsers/mysql"
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
	common.AddUserAgentMetadata("rds", "mysql")

	parser := &mysql.Parser{SampleRate: int(common.GetSampleRate())}
	parser.Init(&mysql.Options{})

	dbh := common.NewMySQLHandler(parser)
	dbh.Env = os.Getenv("ENVIRONMENT")

	if os.Getenv("SCRUB_QUERY") == "true" {
		dbh.ScrubQuery = true
	}

	lambda.Start(dbh.Handle)
}
