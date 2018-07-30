package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestExtractPayload(t *testing.T) {
	line := `{"samplerate":1,"dataset":"travis.serverless-test","data":{"trace.span_id":"28320411-35c9-45dc-b968-b8b9d091f4bf","meta.local_hostname":"ip-10-12-92-115","trace.trace_id":"6614adb4-a74e-47a2-92de-06435f548eb1","name":"sleepytime","service_name":"travis.serverless-test","trace.parent_id":"56c10d68-e8db-46d5-97ad-0d508a3c08bf","duration_ms":500.65700000000004,"meta.beeline_version":"1.0.0"},"user_agent":"libhoney-py/1.4.0","time":"2018-07-23T22:06:58.471593Z"}`
	var data map[string]interface{}
	err := json.Unmarshal([]byte(line), &data)
	if err != nil {
		t.Error("didn't parse json: ", err)
	}
	payload, err := extractPayload(data)
	if err != nil {
		t.Error("extractPayload failed: ", err)
	}
	if payload.dataset != "travis.serverless-test" {
		t.Error("unexpected value for dataset: ", payload.dataset)
	}
	if payload.sampleRate != 1 {
		t.Error("unexpected value for sampleRate: ", payload.sampleRate)
	}
	expectedTime, err := time.Parse("2006-01-02T15:04:05.000000Z", "2018-07-23T22:06:58.471593Z")
	if err != nil {
		t.Error("error parsing time: ", err)
	}
	if expectedTime != payload.time {
		t.Error("unexpected value for time: ", payload.time)
	}
}
