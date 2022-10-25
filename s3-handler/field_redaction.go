package main

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
)

type FieldRedactionConfig struct {
	Name string `json:"name"`
	Repl string `json:"repl"`
	Expr []string `json:"expr"`
}

type fieldRedactor func(string) string

type fieldRedactionHandler struct {
	redactionConfig map[string]fieldRedactor
}

// Nested capture groups are not supported for obvious reasons.
//
func newFieldRedactor(fieldRedactConfig FieldRedactionConfig) fieldRedactor {
	replacement := []byte(fieldRedactConfig.Repl)
	rlen := len(replacement)
	redactors := make([]*regexp.Regexp, len(fieldRedactConfig.Expr))
	for i, expr := range fieldRedactConfig.Expr {
		redactors[i] = regexp.MustCompile(expr)
	}
	return func(fieldValue string) string {
		redacted := []byte(fieldValue)
		for _, redactor := range redactors {
			if matches := redactor.FindAllSubmatchIndex(redacted, -1); matches != nil {
				toRemove := 0
				srcLen := len(redacted)
				for _, loc := range matches {
					for j := 2; j < len(loc); j += 2 {
						toRemove += loc[j+1] - loc[j]
					}
				}
				if toRemove == 0 { // when no capture group is present
					continue
				}
				tmp := make([]byte, 0, srcLen - toRemove + (len(matches) * rlen))
				last := 0
				for _, loc := range matches {
					for j := 2; j < len(loc); j += 2 {
						if loc[j] > last {
							tmp = append(tmp, redacted[last:loc[j]]...)
						}
						tmp = append(tmp, replacement...)
						last = loc[j+1]
					}
				}
				if last < srcLen {
					tmp = append(tmp, redacted[last:]...)
				}
				redacted = tmp
			}
		}
		return string(redacted)
	}
}

func newFieldRedactionHandler(redactionConfigString string) (*fieldRedactionHandler, error) {
	var frs []FieldRedactionConfig
	jsonString, err := base64.StdEncoding.DecodeString(redactionConfigString)
	if err != nil {
		jsonString = []byte(redactionConfigString)
	}
	if err = json.Unmarshal(jsonString, &frs); err != nil {
		return nil, err
	}
  redactionConfig := make(map[string]fieldRedactor, len(frs))
	for _, fr := range frs {
		redactionConfig[fr.Name] = newFieldRedactor(fr)
	}
	return &fieldRedactionHandler{redactionConfig: redactionConfig}, nil
}

func (fr fieldRedactionHandler) redact(line map[string]interface{}) {
	for fieldName, redactor := range fr.redactionConfig {
		if value, ok := line[fieldName]; ok {
			if sValue, typeOk := value.(string); typeOk {
				line[fieldName] = redactor(sValue)
			}
		}
	}
}

