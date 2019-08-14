package common

import (
	"os"
	"testing"
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
