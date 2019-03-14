package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers/postgresql"
	"github.com/honeycombio/libhoney-go"
)

// Response is a simple structured response
type Response struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

var parser *postgresql.Parser
var env, scrubQuery string

// Handler to process Cloudwatch Log events containing PostgreSQL RDS log statements
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

	lines := make(chan string)
	events := make(chan event.Event)
	wg := sync.WaitGroup{}

	go func() {
		for _, event := range data.LogEvents {
			// these are multi log lines, split into individual lines as the
			// parser expects that
			messageLines := strings.Split(event.Message, "\n")
			for _, l := range messageLines {
				lines <- l
			}
		}
		close(lines)
	}()

	go func() {
		wg.Add(1)
		for e := range events {
			hnyEvent := libhoney.NewEvent()
			if scrubQuery == "true" {
				if val, ok := e.Data["query"]; ok {
					// generate a sha256 hash
					newVal := sha256.Sum256([]byte(fmt.Sprintf("%v", val)))
					// and use the base16 string version of it
					e.Data["query"] = fmt.Sprintf("%x", newVal)
				}
			}
			hnyEvent.Add(e.Data)
			hnyEvent.Timestamp = e.Timestamp
			hnyEvent.SampleRate = uint(e.SampleRate)
			if env != "" {
				hnyEvent.AddField("env", env)
			}
			// Sampling is done in the parser for greater efficiency
			hnyEvent.SendPresampled()
		}
		wg.Done()
	}()

	parser.ProcessLines(lines, events, nil)
	close(events)

	wg.Wait()

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

	parser = &postgresql.Parser{}
	parser.Init(&postgresql.Options{})

	common.AddUserAgentMetadata("rds", "postgresql")

	env = os.Getenv("ENVIRONMENT")
	scrubQuery = os.Getenv("SCRUB_QUERY")

	lambda.Start(Handler)
}
