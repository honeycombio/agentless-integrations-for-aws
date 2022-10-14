package common

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/honeytail/parsers/htjson"
	"github.com/honeycombio/honeytail/parsers/keyval"
	"github.com/honeycombio/honeytail/parsers/regex"
	libhoney "github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/transmission"
	"github.com/sirupsen/logrus"
)

var (
	sampleRate   uint
	writeKey     string
	apiHost      string
	dataset      string
	errorDataset string
	filterFields []string
	version      = "dev"
)

// InitHoneycombFromEnvVars will attempt to call libhoney.Init based on values
// passed to the lambda through env vars. The caller is responsible for calling
// libhoney.Close afterward. Will return an err if insufficient ENV vars were
// specified.
func InitHoneycombFromEnvVars() error {
	sampleRate = 1
	if os.Getenv("SAMPLE_RATE") != "" {
		i, err := strconv.Atoi(os.Getenv("SAMPLE_RATE"))
		// including a check on sample rate being configured as 0.
		// assumption: intention of setting 0 is that no sampling is desired.
		if err != nil || uint(i) == 0 {
			logrus.WithField("sample_rate", os.Getenv("SAMPLE_RATE")).
				Warn("Warning: unable to parse sample rate or sample rate configured as 0, falling back to 1.")
		} else {
			sampleRate = uint(i)
		}
	}

	// If KMS_KEY_ID is supplied, we assume we're dealing with an encrypted key.
	kmsKeyID := os.Getenv("KMS_KEY_ID")
	if kmsKeyID != "" {
		encryptedWriteKey := os.Getenv("HONEYCOMB_WRITE_KEY")
		if encryptedWriteKey == "" {
			return fmt.Errorf("Value for KMS_KEY_ID but no value for HONEYCOMB_WRITE_KEY")
		} else {
			kmsSession := session.Must(session.NewSession(&aws.Config{
				Region: aws.String(os.Getenv("AWS_REGION")),
			}))

			config := &aws.Config{}
			svc := kms.New(kmsSession, config)
			cyphertext, err := base64.StdEncoding.DecodeString(encryptedWriteKey)
			if err != nil {
				logrus.WithError(err).
					Error("unable to decode ciphertext in write key")
				return fmt.Errorf("unable to decode ciphertext in write key")
			}
			resp, err := svc.Decrypt(&kms.DecryptInput{
				CiphertextBlob: cyphertext,
			})

			if err != nil {
				logrus.WithError(err).Error("unable to decrypt honeycomb write key")
				return fmt.Errorf("unable to decrypt honeycomb write key")
			}
			writeKey = string(resp.Plaintext)
		}
	} else {
		writeKey = os.Getenv("HONEYCOMB_WRITE_KEY")
		if writeKey == "" {
			return fmt.Errorf("no value for HONEYCOMB_WRITE_KEY")
		}
	}

	apiHost = os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = "https://api.honeycomb.io"
	}

	dataset = os.Getenv("DATASET")
	if dataset == "" {
		dataset = "honeycomb-cloudwatch-logs"
	}

	errorDataset = os.Getenv("ERROR_DATASET")

	libhoney.UserAgentAddition = fmt.Sprintf("integrations-for-aws/%s", version)

	// Call Init to configure libhoney
	libhoney.Init(libhoney.Config{
		WriteKey:   writeKey,
		Dataset:    dataset,
		APIHost:    apiHost,
		SampleRate: sampleRate,
	})

	debug := envOrElseBool("HONEYCOMB_DEBUG", false)
	if debug {
		go readResponses(libhoney.TxResponses())
	}

	return nil
}

// ConstructParser accepts a parser name and attempts to build the parser,
// pulling additional environment variables as needed
func ConstructParser(parserType string) (parsers.LineParser, error) {
	if parserType == "regex" {
		regexVal := os.Getenv("REGEX_PATTERN")
		regexParser, err := regex.NewRegexLineParser([]string{regexVal})
		if err != nil {
			return nil, fmt.Errorf("failed to create regex parser: %s", err.Error())
		}
		return regexParser, nil
	} else if parserType == "json" {
		return &htjson.JSONLineParser{}, nil
	} else if parserType == "keyval" {
		return &keyval.KeyValLineParser{}, nil
	}
	return nil, fmt.Errorf("Unknown parser: %s", parserType)
}

// ConvertTypes will convert strings into integer and floats if applicable
func ConvertTypes(input map[string]interface{}) map[string]interface{} {
	data := make(map[string]interface{})
	for k, v := range input {
		if stringVal, ok := v.(string); ok {
			if val, err := strconv.Atoi(stringVal); err == nil {
				data[k] = val
			} else if val, err := strconv.ParseFloat(stringVal, 64); err == nil {
				data[k] = val
			} else {
				data[k] = stringVal
			}
		} else {
			data[k] = v
		}
	}

	return data
}

// AddUserAgentMetadata adds additional metadata to the user agent string
func AddUserAgentMetadata(handler, parser string) {
	libhoney.UserAgentAddition = fmt.Sprintf(
		"%s (%s, %s)", libhoney.UserAgentAddition, handler, parser,
	)
}

// GetSampleRate returns the sample rate the configured sample rate
func GetSampleRate() uint {
	return sampleRate
}

// WriteErrorEvent writes the error and optional fields to the Error Dataset,
// if an error dataset was specified
func WriteErrorEvent(err error, errorType string, fields map[string]interface{}) {
	if errorDataset != "" {
		ev := libhoney.NewEvent()
		ev.Dataset = errorDataset
		ev.AddField("meta.honeycomb_error", err.Error())
		ev.AddField("meta.error_type", errorType)
		ev.Add(fields)
		ev.Send()
	}
}

// GetFilterFields returns a list of fields to be deleted from an event before it is sent to Honeycomb
// If FILTER_FIELDS is not set, returns an empty list.
func GetFilterFields() []string {
	if filterFields != nil {
		return filterFields
	}

	filterFields = []string{}

	filtersString := os.Getenv("FILTER_FIELDS")
	if filtersString == "" {
		// return an empty (but non-nil) filterField list
		return filterFields
	}

	// FILTER_FIELDS is a comma-separated string of fields
	filterFields = strings.Split(filtersString, ",")

	return filterFields
}

func envOrElseBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		return v
	}
	return fallback
}

func readResponses(responses chan transmission.Response) {
	for r := range responses {
		var metadata string
		if r.Metadata != nil {
			metadata = fmt.Sprintf("%s", r.Metadata)
		}
		if r.StatusCode >= 200 && r.StatusCode < 300 {
			message := "Successfully sent event to Honeycomb"
			if metadata != "" {
				message += fmt.Sprintf(": %s", metadata)
			}
			logrus.Debugf("%s", message)
		} else if r.StatusCode == http.StatusUnauthorized {
			logrus.Errorf("Error sending event to honeycomb! The APIKey was rejected, please verify your APIKey. %s", metadata)
		} else {
			logrus.Errorf("Error sending event to Honeycomb! %s had code %d, err %v and response body %s",
				metadata, r.StatusCode, r.Err, r.Body)
		}
	}
}
