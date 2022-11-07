package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers/mysql"
)

type CloudWatchLogBody struct {
	Type      string               `json:"messageType"`
	Owner     string               `json:"owner"`
	LogGroup  string               `json:"logGroup"`
	LogStream string               `json:"logStream"`
	Events    []CloudWatchLogEvent `json:"logEvents"`
}

type CloudWatchLogEvent struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

var parser *mysql.Parser

func handler(ctx context.Context, input events.KinesisFirehoseEvent) (events.KinesisFirehoseResponse, error) {
	if parser == nil {
		return events.KinesisFirehoseResponse{}, fmt.Errorf("parser not initialized, cannot process events")
	}

	var response events.KinesisFirehoseResponse
	for _, record := range input.Records {
		cwb, err := decodeData(record.Data)
		if err != nil {
			return events.KinesisFirehoseResponse{}, err
		}

		// these messages ensure kinesis can reach the lambda and don't require processing
		if cwb.Type == "CONTROL_MESSAGE" {
			var droppedRecord events.KinesisFirehoseResponseRecord
			droppedRecord.RecordID = record.RecordID
			droppedRecord.Result = events.KinesisFirehoseTransformedStateDropped
			response.Records = append(response.Records, droppedRecord)
		} else if cwb.Type == "DATA_MESSAGE" {
			var parsedEventJson []map[string]interface{}
			for _, v := range cwb.Events {
				lines := make(chan string)
				parsedEvents := make(chan event.Event)
				go func() {
					messageLines := strings.Split(v.Message, "\n")
					for _, l := range messageLines {
						lines <- l
					}
					close(lines)
				}()

				go func() {
					parser.ProcessLines(lines, parsedEvents, nil)
					close(parsedEvents)
				}()

				for p := range parsedEvents {
					hnyEvent := p.Data
					hnyEvent["aws.cloudwatch.owner"] = cwb.Owner
					hnyEvent["aws.cloudwatch.group"] = cwb.LogGroup
					hnyEvent["aws.cloudwatch.stream"] = cwb.LogStream
					hnyEvent["aws.cloudwatch.id"] = v.ID
					hnyEvent["timestamp"] = p.Timestamp
					parsedEventJson = append(parsedEventJson, hnyEvent)
				}
			}

			b, err := json.Marshal(parsedEventJson)
			if err != nil {
				return events.KinesisFirehoseResponse{}, err
			}

			var transformedRecord events.KinesisFirehoseResponseRecord
			transformedRecord.RecordID = record.RecordID
			transformedRecord.Result = events.KinesisFirehoseTransformedStateOk
			transformedRecord.Data = b
			response.Records = append(response.Records, transformedRecord)
		} 
	}
	return response, nil
}

func decodeData(data []byte) (CloudWatchLogBody, error) {
	var cwb CloudWatchLogBody

	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return CloudWatchLogBody{}, err
	}
	gr.Multistream(false)
	defer gr.Close()

	decoder := json.NewDecoder(gr)
	err = decoder.Decode(&cwb)
	if err != nil && err != io.EOF {
		return CloudWatchLogBody{}, err
	}
	return cwb, nil
}

func main() {
	parser = &mysql.Parser{}
	parser.Init(&mysql.Options{})
	common.AddUserAgentMetadata("rds", "mysql")
	lambda.Start(handler)
}
