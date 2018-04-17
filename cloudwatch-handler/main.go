package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
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

	// fetch the write key from the provided SSM key name
	var writeKey string
	writeKeyName := os.Getenv("HONEYCOMB_WRITE_KEY_NAME")
	if writeKeyName == "" {
		log.Printf("Warning: no write key name provided")
	} else {
		ssmSession := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(os.Getenv("AWS_REGION")),
		}))

		config := &aws.Config{}
		svc := ssm.New(ssmSession, config)
		resp, err := svc.GetParameter(&ssm.GetParameterInput{
			Name:           &writeKeyName,
			WithDecryption: aws.Bool(true),
		})

		if err != nil {
			log.Printf("Error: unable to get honeycomb write key from SSM: %s", err.Error())
		}
		writeKey = *resp.Parameter.Value
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
