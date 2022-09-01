package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFilterFields(t *testing.T) {
	fields := GetFilterFields()
	if fields == nil {
		t.Error("GetFilterFields should not return nil")
	}
	if len(fields) != 0 {
		t.Error("GetFilterFields should return an empty slice if FILTER_FIELDS is not set")
	}
	filterFields = nil
	os.Setenv("FILTER_FIELDS", "a,b,c")
	fields = GetFilterFields()
	if len(fields) != 3 {
		t.Error("expected GetFilterFields to return 3 strings")
	}
	if fields[0] != "a" {
		t.Error("wrong value in GetFilterFields result")
	}
	if fields[1] != "b" {
		t.Error("wrong value in GetFilterFields result")
	}
	if fields[2] != "c" {
		t.Error("wrong value in GetFilterFields result")
	}
}

func TestGetRenameFields(t *testing.T) {
	ClearCache()

	t.Setenv("RENAME_FIELDS", "trace_id=trace.trace_id,span_id=trace.span_id")

	expected := map[string]string{"trace_id": "trace.trace_id", "span_id": "trace.span_id"}
	assert.Equal(t, expected, GetRenameFields())

	t.Setenv("RENAME_FIELDS", "cached=value")
	assert.Equal(t, expected, GetRenameFields())
}

func TestGetAliasFields(t *testing.T) {
	ClearCache()

	t.Setenv("ALIAS_FIELDS", "trace_id=trace.trace_id,span_id=trace.span_id")

	expected := map[string]string{"trace_id": "trace.trace_id", "span_id": "trace.span_id"}
	assert.Equal(t, expected, GetAliasFields())

	t.Setenv("ALIAS_FIELDS", "cached=value")
	assert.Equal(t, expected, GetAliasFields())
}
