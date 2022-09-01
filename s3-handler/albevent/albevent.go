package albevent

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/libhoney-go"
)

type Event struct {
	mappings map[string]interface{}
	event    *libhoney.Event

	Timestamp time.Time
}

func (e *Event) Fields() map[string]interface{} {
	return e.event.Fields()
}

func (e *Event) Field(name string) interface{} {
	return e.event.Fields()[name]
}

func (e *Event) Event() *libhoney.Event {
	return e.event
}

func (e *Event) AddField(name string, value interface{}) {
	e.event.AddField(name, value)
}

func (e *Event) AddFields(values map[string]interface{}) {
	e.event.Add(values)
}

func (e *Event) handleTimestamps() {
	timeFieldName := os.Getenv("TIME_FIELD_NAME")
	timeFieldFormat := os.Getenv("TIME_FIELD_FORMAT")

	e.event.Timestamp = httime.GetTimestamp(e.mappings, timeFieldName, timeFieldFormat)
	e.Timestamp = e.event.Timestamp

	if seconds, err := strconv.ParseFloat(fmt.Sprintf("%s", e.mappings["target_processing_time"]), 32); err == nil {
		e.event.AddField("duration_ms", float32(seconds*1000))
	}
}

func (e *Event) addParsedAttributes() {
	if os.Getenv("PARSER_TYPE") != "json" {
		e.event.Add(common.ConvertTypes(e.mappings))
	} else {
		e.event.Add(e.mappings)
	}
}

func (e *Event) addTraceId() {
	if "w3c" != os.Getenv("TRACE_FORMAT") {
		if e.mappings["trace_id"] == nil {
			return
		}

		e.event.AddField("trace.trace_id", e.mappings["trace_id"])
		e.event.AddField("trace.span_id", e.mappings["trace_id"])
		return
	}

	if traceParts, err := getW3CTraceParts(e.mappings["trace_id"]); err == nil {
		e.event.AddField("trace.trace_id", traceParts.traceId)
		e.event.AddField("trace.span_id", traceParts.spanId)
	}
}

func getW3CTraceParts(traceId interface{}) (*TraceInfo, error) {
	amzTraceId := fmt.Sprintf("%s", traceId)

	if len(amzTraceId) < 35 {
		return nil, errors.New(fmt.Sprintf("trace id (%s) is in an invalid format", amzTraceId))
	}

	w3cTraceId := amzTraceId[2:10] + amzTraceId[11:]
	w3cSpanId := amzTraceId[len(amzTraceId)-16:]

	return &TraceInfo{traceId: w3cTraceId, spanId: w3cSpanId}, nil
}

func (e *Event) addUrlAttributes() *Event {
	if urlInfo, err := getUrlInfo(e.mappings["request"]); err == nil {
		e.AddFields(map[string]interface{}{
			"name":                 fmt.Sprintf("%s %s", urlInfo.Method, urlInfo.Path),
			"request.method":       urlInfo.Method,
			"request.path":         urlInfo.Path,
			"request.query_string": urlInfo.QueryString,
		})
	}

	return e
}

func (e *Event) inspectErrorCause() *Event {
	if error_cause, ok := e.mappings["error_cause"]; ok && error_cause != "-" {
		e.AddField("error", error_cause)
	}

	return e
}

func (e *Event) Send() {
	e.event.Send()
}

type TraceInfo struct {
	traceId string
	spanId  string
}

type UrlInfo struct {
	Method      string
	Path        string
	QueryString string
}

func getUrlInfo(requestValue interface{}) (*UrlInfo, error) {
	request := fmt.Sprintf("%s", requestValue)
	requestParts := strings.Split(fmt.Sprintf("%s", request), " ")
	if len(requestParts) != 3 {
		return nil, errors.New(fmt.Sprintf("request (%s) in invalid format", request))
	}

	url, err := url.Parse(requestParts[1])
	if err != nil {
		return nil, err
	}

	return &UrlInfo{
		Method:      requestParts[0],
		Path:        url.Path,
		QueryString: url.RawQuery,
	}, nil
}

func filterFields(mappings map[string]interface{}) {
	for _, field := range common.GetFilterFields() {
		delete(mappings, field)
	}
}

func renameFields(mappings map[string]interface{}) {
	for k, v := range common.GetRenameFields() {
		if tmp, ok := mappings[k]; ok {
			mappings[v] = tmp
			delete(mappings, k)
		}
	}
}

func aliasFields(mappings map[string]interface{}) {
	for k, v := range common.GetAliasFields() {
		if tmp, ok := mappings[k]; ok {
			mappings[v] = tmp
		}
	}
}

func NewEvent(mappings map[string]interface{}) Event {
	filterFields(mappings)
	renameFields(mappings)
	aliasFields(mappings)

	event := libhoney.NewEvent()
	albEvent := Event{mappings: mappings, event: event}
	albEvent.AddFields(map[string]interface{}{"service_name": "alb", "environment": os.Getenv("ENVIRONMENT")})

	albEvent.handleTimestamps()
	albEvent.addParsedAttributes()
	albEvent.addTraceId()
	albEvent.addUrlAttributes()
	albEvent.inspectErrorCause()

	return albEvent
}
