package common

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/events"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/honeytail/parsers/mysql"
	"github.com/honeycombio/honeytail/parsers/postgresql"
	libhoney "github.com/honeycombio/libhoney-go"
)

// Response is a simple structured response
type Response struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

type DBHandler struct {
	parser parsers.Parser

	// presampledRate holds a non-zero value if the events are presampled.
	//
	// This is necessary because the MySQL parser will presample, but the
	// Postgres one will not. We need to know the difference to be able to set
	// the sample rate on events and call Send() vs SendPresampled().
	presampledRate uint

	ScrubQuery bool
	Env        string
}

// NewMySQLHandler creates a DBHandler with presampledRate set correctly.
func NewMySQLHandler(parser *mysql.Parser) *DBHandler {
	return &DBHandler{
		parser:         parser,
		presampledRate: uint(parser.SampleRate),
	}
}

// NewPostgreSQLHandler returns a DBHandler.
func NewPostgreSQLHandler(parser *postgresql.Parser) *DBHandler {
	return &DBHandler{
		parser: parser,
	}
}

func (h *DBHandler) Handle(request events.CloudwatchLogsEvent) (Response, error) {
	if h.parser == nil {
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
			if h.ScrubQuery {
				if val, ok := e.Data["query"]; ok {
					// generate a sha256 hash
					newVal := sha256.Sum256([]byte(fmt.Sprintf("%v", val)))
					// and use the base16 string version of it
					e.Data["query"] = fmt.Sprintf("%x", newVal)
				}
			}
			hnyEvent.Add(e.Data)
			hnyEvent.Timestamp = e.Timestamp
			if h.Env != "" {
				hnyEvent.AddField("env", h.Env)
			}
			fields := hnyEvent.Fields()
			for _, field := range GetFilterFields() {
				delete(fields, field)
			}
			// Add CloudWatch event metadata
			hnyEvent.AddField("aws.cloudwatch.group", data.LogGroup)
			hnyEvent.AddField("aws.cloudwatch.stream", data.LogStream)
			hnyEvent.AddField("aws.cloudwatch.owner", data.Owner)

			// MySQL sampling is done in the parser
			if h.presampledRate > 0 {
				hnyEvent.SampleRate = h.presampledRate
				hnyEvent.SendPresampled()
			} else {
				hnyEvent.Send()
			}
		}
		wg.Done()
	}()

	h.parser.ProcessLines(lines, events, nil)
	close(events)

	wg.Wait()

	libhoney.Flush()

	return Response{
		Ok:      true,
		Message: "ok",
	}, nil
}
