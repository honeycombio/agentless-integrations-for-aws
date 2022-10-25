package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const RedactionConfigJSON = `[
	{
		"name": "request",
		"repl": "[REDACTED]",
		"expr": [
			"([a-zA-Z0-9\\.!\\#\\$%*+/\\^_\\{\\|\\}\\~\\-]+(?:@|%40)[a-zA-Z0-9][a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]{2,})*\\.[a-zA-Z]{2,})",
			"address=\"([^\"]+)"
		]
	},
	{
		"name": "other_field",
		"repl": "[REDACTED_OTHER]",
		"expr": ["foo=\\\\?\"(bar)\\\\?\""]
	}
]`

const RedactionConfigB64 = "WwogIHsKICAgICJuYW1lIjogInJlcXVlc3QiLAogICAgInJlcGwiOiAiW1JFREFDVEVEXzFdIiwKICAgICJleHByIjogWwogICAgICAgICAgICAiKFthLXpBLVowLTlcXC4hXFwjXFwkJVxcJuKAmSorLz1cXD9cXF5fXFx7XFx8XFx9XFx+XFwtXSsoPzpAfCU0MClbYS16QS1aMC05XVthLXpBLVowLTktXSsoPzpcXC5bYS16QS1aMC05LV17Mix9KSpcXC5bYS16QS1aXXsyLH0pIiwKICAgICAgICAgICAgImFkZHJlc3M9XCIoW15cIl0rKSIKICAgIF0KICB9LAogIHsKICAgICJuYW1lIjogIm90aGVyX2ZpZWxkIiwKICAgICJyZXBsIjogIltSRURBQ1RFRF8yXSIsCiAgICAiZXhwciI6IFsiZm9vPVwiKGJcXC5yKVwiIl0KICB9Cl0K"

func TestFieldRedaction(t *testing.T) {
	handler, err := newFieldRedactionHandler(RedactionConfigJSON)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(handler.redactionConfig))

	testLine := map[string]interface{} {
		"request": `/path?user=test.user@example.com&address="mystreet 12 12345 city"&key=value&email="test.user@example.com"`,
		"other_field": `query="foo=\"bar\"&key=value"`,
		"ignored_field" : `ignore=test.user@example.com&foo=\"bar\"`,
	}

	handler.redact(testLine)
	assert.Equal(t, `/path?user=[REDACTED]&address="[REDACTED]"&key=value&email="[REDACTED]"`, testLine["request"].(string))
	assert.Equal(t, `query="foo=\"[REDACTED_OTHER]\"&key=value"`, testLine["other_field"].(string))
	assert.Equal(t, `ignore=test.user@example.com&foo=\"bar\"`, testLine["ignored_field"].(string))
}

func TestFieldRedactionBase64(t *testing.T) {
	handler, err := newFieldRedactionHandler(RedactionConfigB64)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(handler.redactionConfig))
}
