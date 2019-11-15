package main

import (
	"fmt"
	"time"

	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/libhoney-go"
)

type payload struct {
	time       time.Time
	sampleRate uint
	dataset    string
	data       interface{}
}

func extractPayload(data map[string]interface{}) (*payload, error) {
	p := &payload{}
	_data, ok := data["data"]
	if !ok {
		return nil, fmt.Errorf("unable to find data in payload")
	}
	if timestamp, ok := data["time"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, timestamp); err == nil {
			p.time = parsedTime
		}
	}
	if dataset, ok := data["dataset"].(string); ok {
		p.dataset = dataset
	}
	if sampleRate, ok := data["samplerate"].(float64); ok {
		p.sampleRate = uint(sampleRate)
	}
	p.data = _data

	return p, nil
}

// Response is a simple structured response
type Response struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

var parser parsers.LineParser
var parserType, timeFieldName, timeFieldFormat, env string

func Handler(request events.CloudwatchLogsEvent) (Response, error) {
	if parser == nil {
		return Response{
			Ok:      false,
			Message: "parser not initialized, cannot process events",
		}, fmt.Errorf("parser not initialized, cannot process events")
	}

	data, err := request.AWSLogs.Parse()
	if err != nil {
		return Response{
			Ok:      false,
			Message: fmt.Sprintf("failed to parse cloudwatch event data: %s", err.Error()),
		}, err
	}

	for _, event := range data.LogEvents {
		parsedLine, err := parser.ParseLine(event.Message)
		if err != nil {
			logrus.WithError(err).WithField("line", event.Message).
				Warn("unable to parse line, skipping")
			continue
		}
		// The JSON parser returns a map[string]interface{} - we need to convert it
		// to a structure we can work with
		payload, err := extractPayload(parsedLine)
		if err != nil {
			logrus.WithError(err).WithField("line", event.Message).
				Warn("unable to get event payload from line, skipping")
			continue
		}
		hnyEvent := libhoney.NewEvent()
		// add the actual event data
		hnyEvent.Add(payload.data)
		// Include the logstream that this data came from to make it easier to find the source
		// in Cloudwatch
		hnyEvent.AddField("aws.cloudwatch.logstream", data.LogStream)

		// If we have sane values for other fields, set those as well
		if !payload.time.IsZero() {
			hnyEvent.Timestamp = payload.time
		}
		if payload.dataset != "" {
			hnyEvent.Dataset = payload.dataset
		}
		if payload.sampleRate > 0 {
			hnyEvent.SampleRate = payload.sampleRate
		}

		// We don't sample here - we assume it has been done upstream by
		// whatever wrote to the log
		hnyEvent.SendPresampled()
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

	parser, err = common.ConstructParser("json")
	if err != nil {
		logrus.WithError(err).WithField("parser_type", parserType).
			Fatal("unable to construct parser")
		return
	}
	common.AddUserAgentMetadata("publisher", "json")

	lambda.Start(Handler)
}
