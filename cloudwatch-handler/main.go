package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/honeycombio/honeytail/parsers/keyval"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/honeytail/parsers/htjson"
	"github.com/honeycombio/honeytail/parsers/regex"
	"github.com/honeycombio/libhoney-go"
)

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
			log.Printf("error parsing line: %s - line was `%s`", err, event.Message)
			continue
		}
		hnyEvent := libhoney.NewEvent()

		timestamp := httime.GetTimestamp(parsedLine, timeFieldName, timeFieldFormat)
		hnyEvent.Timestamp = timestamp

		// convert ints and floats if necessary
		if parserType != "json" {
			data := make(map[string]interface{})
			for k, v := range parsedLine {
				if stringVal, ok := v.(string); ok {
					if val, err := strconv.Atoi(stringVal); err == nil {
						data[k] = val
					} else if val, err := strconv.ParseFloat(stringVal, 64); err == nil {
						data[k] = val
					} else {
						data[k] = stringVal
					}
				} else {
					data[k] = v
				}
			}
			hnyEvent.Add(data)
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
		return &htjson.JSONLineParser{}
	} else if parserType == "keyval" {
		return &keyval.KeyValLineParser{}
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

	parserType = os.Getenv("PARSER_TYPE")
	parser = constructParser(parserType)

	// attempt to decrypt the write key provided
	var writeKey string
	encryptedWriteKey := os.Getenv("HONEYCOMB_WRITE_KEY")
	if encryptedWriteKey == "" {
		log.Printf("Warning: no write key provided")
	} else {
		kmsSession := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(os.Getenv("AWS_REGION")),
		}))

		config := &aws.Config{}
		svc := kms.New(kmsSession, config)
		cyphertext, err := base64.StdEncoding.DecodeString(encryptedWriteKey)
		if err != nil {
			log.Printf("error decoding ciphertext in write key: %s", err.Error())
		}
		resp, err := svc.Decrypt(&kms.DecryptInput{
			CiphertextBlob: cyphertext,
		})

		if err != nil {
			log.Printf("Error: unable to decrypt honeycomb write key: %s", err.Error())
		}
		writeKey = string(resp.Plaintext)
	}

	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = "https://api.honeycomb.io"
	}

	dataset := os.Getenv("DATASET")
	if dataset == "" {
		dataset = "honeycomb-cloudwatch-logs"
	}

	env = os.Getenv("ENVIRONMENT")
	timeFieldName = os.Getenv("TIME_FIELD_NAME")
	timeFieldFormat = os.Getenv("TIME_FIELD_FORMAT")

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
