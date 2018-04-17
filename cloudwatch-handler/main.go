package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/honeytail/parsers/regex"
	"github.com/honeycombio/libhoney-go"
)

// Response is a simple structured response
type Response struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

var parser parsers.LineParser

//(?P<version>\d+) (?P<account_id>\d+) (?P<interface_id>eni-[0-9a-f]+) (?P<src_addr>[\d\.]+) (?P<dst_addr>[\d\.]+) (?P<src_port>\d+) (?P<dst_port>\d+) (?P<protocol>\d+) (?P<packets>\d+) (?P<bytes>\d+) (?P<start_time>\d+) (?P<end_time>\d+) (?P<action>[A-Z]+) (?P<log_status>[A-Z]+)
// Handler is your Lambda function handler
// It uses Amazon API Gateway request/responses provided by the aws-lambda-go/events package,
// However you could use other event sources (S3, Kinesis etc), or JSON-decoded primitive types such as 'string'.
func Handler(request events.CloudwatchLogsEvent) (Response, error) {
	env := os.Getenv("ENVIRONMENT")

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
		parsedData, err := parser.ParseLine(event.Message)
		if err != nil {
			log.Printf("error parsing line: %s - line was `%s`", err, event.Message)
			continue
		}

		hnyEvent := libhoney.NewEvent()
		hnyEvent.Add(parsedData)
		hnyEvent.AddField("env", env)
		hnyEvent.Send()
	}

	return Response{
		Ok:      true,
		Message: "ok",
	}, nil
}

func constructParser(parserType string) parsers.LineParser {
	if parserType == "regex" {
		regexVal := os.Getenv("REGEX_PATTERN")
		regexParser, err := regex.NewRegexLineParser([]string{regexVal})
		if err != nil {
			log.Printf("Error: failed to construct regex parser")
			return nil
		}
		return regexParser
	} else if parserType == "json" {
		return nil
	}
	return nil
}

func main() {
	var sampleRate uint = 1
	if os.Getenv("SAMPLE_RATE") != "" {
		i, err := strconv.Atoi(os.Getenv("SAMPLE_RATE"))
		if err != nil {
			log.Printf("Warning: unable to parse sample rate %s, falling back to 1.",
				os.Getenv("SAMPLE_RATE"))
		}
		sampleRate = uint(i)
	}

	parserType := os.Getenv("PARSER_TYPE")
	parser = constructParser(parserType)

	writeKey := os.Getenv("HONEYCOMB_WRITE_KEY")
	if writeKey == "" {
		log.Printf("Warning: no write key set")
	}

	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = "https://api.honeycomb.io"
	}

	dataset := os.Getenv("DATASET")
	if dataset == "" {
		dataset = "honeycomb-cloudwatch-logs"
	}

	// Call Init to configure libhoney
	libhoney.Init(libhoney.Config{
		WriteKey:   writeKey,
		Dataset:    dataset,
		APIHost:    apiHost,
		SampleRate: sampleRate,
	})
	defer libhoney.Close()
	lambda.Start(Handler)
}
