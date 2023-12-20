package main

import (
	"context"
	"encoding/json"

	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type snsEntityForHny struct {
	events.SNSEntity
	Message map[string]interface{}
}

func failedRecord(in events.KinesisFirehoseEventRecord) events.KinesisFirehoseResponseRecord {
	var failedRecord events.KinesisFirehoseResponseRecord
	failedRecord.RecordID = in.RecordID
	failedRecord.Result = events.KinesisFirehoseTransformedStateProcessingFailed
	failedRecord.Data = in.Data

	return failedRecord
}

func handler(ctx context.Context, input events.KinesisFirehoseEvent) (events.KinesisFirehoseResponse, error) {
	var response events.KinesisFirehoseResponse
	for _, record := range input.Records {
		var snsEntity events.SNSEntity
		err := json.Unmarshal(record.Data, &snsEntity)
		if err != nil {
			logrus.WithError(err).Warn("failed unmarshal outer event")

			response.Records = append(response.Records, failedRecord(record))
			continue // keep processing other events
		}

		var hnyEvent = snsEntityForHny{snsEntity, map[string]interface{}{}}
		err = json.Unmarshal([]byte(snsEntity.Message), &hnyEvent.Message)
		if err != nil {
			logrus.WithError(err).Warn("failed unmarshal snsEntity.Message")
			// this is not a failed event; the outer record might still be useful
			hnyEvent.Message["rawMessage"] = snsEntity.Message
		}

		b, err := json.Marshal(hnyEvent)
		if err != nil {
			logrus.WithError(err).Warn("couldn't marshal json")

			response.Records = append(response.Records, failedRecord(record))
			continue // keep processing other events
		}

		var transformedRecord events.KinesisFirehoseResponseRecord
		transformedRecord.RecordID = record.RecordID
		transformedRecord.Result = events.KinesisFirehoseTransformedStateOk
		transformedRecord.Data = b
		response.Records = append(response.Records, transformedRecord)
	}

	return response, nil
}

func main() {
	common.AddUserAgentMetadata("sns-kfh-handler", "json")
	lambda.Start(handler)
}
