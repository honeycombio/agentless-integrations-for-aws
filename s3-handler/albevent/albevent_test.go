package albevent

import (
	"fmt"
	"testing"
	"time"

	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/stretchr/testify/assert"
)

var (
	timestampLayout string = "2006-01-02T15:04:05.9999Z"
)

func TestAlbEventTimestamp(t *testing.T) {
	t.Setenv("TIME_FIELD_NAME", "fancy_timestamp")
	t.Setenv("TIME_FIELD_FORMAT", timestampLayout)

	mappings := map[string]interface{}{"fancy_timestamp": "2022-01-07T13:12:11.1234Z"}
	albEvent := NewEvent(mappings)

	expected, _ := time.Parse(timestampLayout, "2022-01-07T13:12:11.1234Z")

	assert.Equal(t, expected, albEvent.Timestamp, "it gets its timestamp from the mappings")
}

func TestItSetsTheServiceName(t *testing.T) {
	albEvent := NewEvent(map[string]interface{}{})
	actual := albEvent.Field("service_name")

	assert.Equal(t, "alb", actual)
}

func TestItSetsTheEnvironment(t *testing.T) {
	t.Setenv("ENVIRONMENT", "envy")

	albEvent := NewEvent(map[string]interface{}{})
	actual := albEvent.Field("environment")

	assert.Equal(t, "envy", actual)
}

func TestItPassesThroughAmznTraceIdByDefault(t *testing.T) {
	amznTraceEpoch := "abcdef12"              // 8-hex epoch
	amznTraceId := "abc123def456ghi789jkl012" // 24-hex unique trace id
	xAmznTraceId := fmt.Sprintf("1-%s-%s", amznTraceEpoch, amznTraceId)

	mappings := map[string]interface{}{"trace_id": xAmznTraceId}
	albEvent := NewEvent(mappings)

	assert.Equal(t, xAmznTraceId, albEvent.Field("trace.trace_id"))

	assert.Equal(t, xAmznTraceId, albEvent.Field("trace.span_id"))
}

func TestItUsesW3CTraceIdsIfAsked(t *testing.T) {
	t.Setenv("TRACE_FORMAT", "w3c")

	amznTraceEpoch := "abcdef12"              // 8-hex epoch
	amznTraceId := "abc123def456ghi789jkl012" // 24-hex unique trace id

	mappings := map[string]interface{}{"trace_id": fmt.Sprintf("1-%s-%s", amznTraceEpoch, amznTraceId)}
	albEvent := NewEvent(mappings)

	assert.Equal(t, amznTraceEpoch+amznTraceId, albEvent.Field("trace.trace_id"))

	expectedSpanId := amznTraceId[len(amznTraceId)-16:]
	assert.Equal(t, expectedSpanId, albEvent.Field("trace.span_id"))
}

func TestItDerivesItsDurationFromTargetProcessingTime(t *testing.T) {
	mappings := map[string]interface{}{"target_processing_time": "0.1234"}
	albEvent := NewEvent(mappings)
	fields := albEvent.Fields()

	var expected float32 = 123.4
	assert.Equal(t, expected, fields["duration_ms"])
}

func TestItAddsRequestInfo(t *testing.T) {
	request := "GET https://whatever.com:443/path/to/thing.html?moar=stuff HTTP/2.0"
	mappings := map[string]interface{}{"request": request}
	albEvent := NewEvent(mappings)

	fields := albEvent.Fields()

	assert.Equal(t, "GET /path/to/thing.html", fields["name"])
	assert.Equal(t, "GET", fields["request.method"])
	assert.Equal(t, "/path/to/thing.html", fields["request.path"])
	assert.Equal(t, "moar=stuff", fields["request.query_string"])
}

func TestItAddsTheParsedValuesToTheTrace(t *testing.T) {
	mappings := map[string]interface{}{"what": "ever", "things": 123}
	albEvent := NewEvent(mappings)

	fields := albEvent.Fields()

	assert.Equal(t, "ever", fields["what"])
	assert.Equal(t, 123, fields["things"])
}

func TestItHandlesAlbErrorCauses(t *testing.T) {
	tests := []struct {
		error_cause    string
		expected_error interface{}
	}{
		{error_cause: "-", expected_error: nil},
		{error_cause: "Bad Stuff", expected_error: "Bad Stuff"},
	}

	for _, tt := range tests {
		t.Run(tt.error_cause, func(t *testing.T) {
			mappings := map[string]interface{}{"error_cause": tt.error_cause}
			albEvent := NewEvent(mappings)

			assert.Equal(t, tt.expected_error, albEvent.Field("error"))
		})
	}
}

func TestItCanAddFields(t *testing.T) {
	mappings := map[string]interface{}{}
	albEvent := NewEvent(mappings)

	albEvent.AddField("some", "value")
	albEvent.AddField("other_stuff", float32(1.23))

	assert.Equal(t, "value", albEvent.Field("some"))
	assert.Equal(t, float32(1.23), albEvent.Field("other_stuff"))
}

func TestItCanAddMultipleFields(t *testing.T) {
	mappings := map[string]interface{}{}
	albEvent := NewEvent(mappings)

	albEvent.AddFields(map[string]interface{}{"flerpn": "derpn", "hi": 123})

	assert.Equal(t, "derpn", albEvent.Field("flerpn"))
	assert.Equal(t, 123, albEvent.Field("hi"))
}

func TestItCanFilterFields(t *testing.T) {
	t.Setenv("FILTER_FIELDS", "no,remove_me")

	common.ClearFieldMappingCache()

	mappings := map[string]interface{}{"yes": "keep", "no": "bad", "remove_me": false}
	albEvent := NewEvent(mappings)

	expected := map[string]interface{}{"yes": "keep", "service_name": "alb", "environment": ""}
	assert.Equal(t, expected, albEvent.Fields())
}

func TestItCanRenameFields(t *testing.T) {
	t.Setenv("RENAME_FIELDS", "some_field=new_jack")

	common.ClearFieldMappingCache()

	mappings := map[string]interface{}{"some_field": "whatever"}
	albEvent := NewEvent(mappings)

	assert.Equal(t, "whatever", albEvent.Field("new_jack"))
	assert.Equal(t, nil, albEvent.Field("some_field"))
}

func TestItCanAliasFields(t *testing.T) {
	t.Setenv("ALIAS_FIELDS", "keep_me=new_jack")

	common.ClearFieldMappingCache()

	mappings := map[string]interface{}{"keep_me": "whatever"}
	albEvent := NewEvent(mappings)

	assert.Equal(t, "whatever", albEvent.Field("new_jack"))
	assert.Equal(t, "whatever", albEvent.Field("keep_me"))
}
