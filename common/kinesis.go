package common

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
)

type EventProcessFunc func(lines <-chan string, send chan<- event.Event, prefixRegex *parsers.ExtRegexp)

func decodeKinesisFirehoseData(data []byte) (events.CloudwatchLogsData, error) {
	var cwb events.CloudwatchLogsData

	// kinesis firehose send data payloads gzipped
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return events.CloudwatchLogsData{}, err
	}
	gr.Multistream(false)
	defer gr.Close()

	decoder := json.NewDecoder(gr)
	err = decoder.Decode(&cwb)
	if err != nil && err != io.EOF {
		return events.CloudwatchLogsData{}, err
	}
	return cwb, nil
}

func ProcessKinesisFirehoseEvent(e events.KinesisFirehoseEvent, processFunc EventProcessFunc) (events.KinesisFirehoseResponse, error) {
	var response events.KinesisFirehoseResponse
	for _, record := range e.Records {
		cwb, err := decodeKinesisFirehoseData(record.Data)
		if err != nil {
			// not CloudWatch Logs data? Just put it back on the stream untouched
			var unknownRecord events.KinesisFirehoseResponseRecord
			unknownRecord.RecordID = record.RecordID
			unknownRecord.Result = events.KinesisFirehoseTransformedStateOk
			unknownRecord.Data = record.Data
			response.Records = append(response.Records, unknownRecord)
			continue
		}

		// these messages ensure kinesis can reach the lambda and don't require processing
		if cwb.MessageType == "CONTROL_MESSAGE" {
			var droppedRecord events.KinesisFirehoseResponseRecord
			droppedRecord.RecordID = record.RecordID
			droppedRecord.Result = events.KinesisFirehoseTransformedStateDropped
			response.Records = append(response.Records, droppedRecord)
		} else if cwb.MessageType == "DATA_MESSAGE" {
			var parsedEventJson []map[string]interface{}
			for _, v := range cwb.LogEvents {
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
					processFunc(lines, parsedEvents, nil)
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
				// put the record back on the stream as a processing failure
				var failedRecord events.KinesisFirehoseResponseRecord
				failedRecord.RecordID = record.RecordID
				failedRecord.Result = events.KinesisFirehoseTransformedStateProcessingFailed
				failedRecord.Data = record.Data
				response.Records = append(response.Records, failedRecord)
				continue
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
