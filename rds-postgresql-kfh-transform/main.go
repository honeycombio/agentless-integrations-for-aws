package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/honeycombio/honeytail/parsers/postgresql"
)

var parser *postgresql.Parser

func handler(ctx context.Context, input events.KinesisFirehoseEvent) (events.KinesisFirehoseResponse, error) {
	if parser == nil {
		return events.KinesisFirehoseResponse{}, fmt.Errorf("parser not initialized, cannot process events")
	}

	return common.ProcessKinesisFirehoseEvent(input, parser.ProcessLines)
}

func main() {
	parser = &postgresql.Parser{}
	parser.Init(&postgresql.Options{LogLinePrefix: "%t:%r:%u@%d:[%p]:"})
	common.AddUserAgentMetadata("rds", "postgresql")
	lambda.Start(handler)
}
