package main

import (
	"fmt"
	"os"

	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/libhoney-go"
)

// Response is a simple structured response
type Response struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

var parser parsers.LineParser
var parserType, timeFieldName, timeFieldFormat, env string

func Handler(request events.SNSEvent) (Response, error) {
	if parser == nil {
		return Response{
			Ok:      false,
			Message: "parser not initialized, cannot process events",
		}, fmt.Errorf("parser not initialized, cannot process events")
	}

	for _, event := range request.Records {
		parsedLine, err := parser.ParseLine(event.SNS.Message)
		if err != nil {
			logrus.WithError(err).WithField("line", event.SNS.Message).
				Warn("unable to parse line")
			continue
		}
		hnyEvent := libhoney.NewEvent()

		timestamp := httime.GetTimestamp(parsedLine, timeFieldName, timeFieldFormat)
		hnyEvent.Timestamp = timestamp

		// convert ints and floats if necessary
		if parserType != "json" {
			hnyEvent.Add(common.ConvertTypes(parsedLine))
		} else {
			hnyEvent.Add(parsedLine)
		}

		hnyEvent.AddField("env", env)
		hnyEvent.Send()
	}

	libhoney.Flush()

	return Response{
		Ok:      true,
		Message: "ok",
	}, nil
}

func main() {
	var err error
	if err = common.InitHoneycombFromEnvVars(); err != nil {
		logrus.WithError(err).
			Fatal("Unable to initialize libhoney with the supplied environment variables")
		return
	}
	defer libhoney.Close()

	parserType = os.Getenv("PARSER_TYPE")
	parser, err = common.ConstructParser(parserType)
	if err != nil {
		logrus.WithError(err).WithField("parser_type", parserType).
			Fatal("unable to construct parser")
		return
	}
	common.AddUserAgentMetadata("sns", parserType)

	env = os.Getenv("ENVIRONMENT")
	timeFieldName = os.Getenv("TIME_FIELD_NAME")
	timeFieldFormat = os.Getenv("TIME_FIELD_FORMAT")

	lambda.Start(Handler)
}
