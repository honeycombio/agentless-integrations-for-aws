package main

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/honeycombio/agentless-integrations-for-aws/common"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
)

// Response is a simple structured response
type Response struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

var parser parsers.LineParser
var parserType, timeFieldName, timeFieldFormat, env string
var matchPatterns, filterPatterns []string
var bufferSize uint
var forceGunzip bool
var renameFields = map[string]string{}

func Handler(request events.S3Event) (Response, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	config := &aws.Config{}
	svc := s3.New(sess, config)

	for _, record := range request.Records {
		if filterKey(record.S3.Object.Key, matchPatterns, filterPatterns) {
			logrus.WithFields(logrus.Fields{
				"key":             record.S3.Object.Key,
				"match_patterns":  matchPatterns,
				"filter_patterns": filterPatterns,
			}).Info("key doesn't match specified patterns, skipping")
			continue
		}

		reader, err := getReaderForKey(svc, record.S3.Bucket.Name, record.S3.Object.Key)
		if err != nil {
			continue
		}

		linesRead := 0
		scanner := bufio.NewScanner(reader)
		buffer := make([]byte, bufferSize)
		scanner.Buffer(buffer, int(bufferSize))
		ok := scanner.Scan()
		for ok {
			linesRead++
			if linesRead%10000 == 0 {
				logrus.WithFields(logrus.Fields{
					"lines_read": linesRead,
					"key":        record.S3.Object.Key,
				}).Info("parser checkpoint")
			}
			parsedLine, err := parser.ParseLine(scanner.Text())
			if err != nil {
				logrus.WithError(err).WithField("line", scanner.Text()).
					Warn("failed to parse line")
				common.WriteErrorEvent(err, "parse error", map[string]interface{}{
					"meta.raw_message": scanner.Text(),
				})
				continue
			}

			for k, v := range renameFields {
				if tmp, ok := parsedLine[k]; ok {
					parsedLine[v] = tmp
					delete(parsedLine, k)
				}
			}

			hnyEvent := libhoney.NewEvent()

			timestamp := httime.GetTimestamp(parsedLine, timeFieldName, timeFieldFormat)
			hnyEvent.Timestamp = timestamp

			// convert ints and floats if necessary
			if parserType != "json" {
				hnyEvent.Add(common.ConvertTypes(parsedLine))
			} else {
				hnyEvent.Add(parsedLine)
			}

			hnyEvent.AddField("aws.s3.bucket", record.S3.Bucket.Name)
			hnyEvent.AddField("aws.s3.object", record.S3.Object.Key)
			hnyEvent.AddField("env", env)
			fields := hnyEvent.Fields()
			for _, field := range common.GetFilterFields() {
				delete(fields, field)
			}
			hnyEvent.Send()
			ok = scanner.Scan()
		}

		if scanner.Err() != nil {
			logrus.WithError(scanner.Err()).WithField("key", record.S3.Object.Key).
				Error("s3 read of object ended early due to error")
		}
	}

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

	parserType = os.Getenv("PARSER_TYPE")
	parser, err = common.ConstructParser(parserType)
	if err != nil {
		logrus.WithError(err).WithField("parser_type", parserType).
			Fatal("unable to construct parser")
		return
	}
	common.AddUserAgentMetadata("s3", parserType)

	env = os.Getenv("ENVIRONMENT")
	timeFieldName = os.Getenv("TIME_FIELD_NAME")
	timeFieldFormat = os.Getenv("TIME_FIELD_FORMAT")

	matchPatterns = []string{".*"}
	filterPatterns = []string{}
	bufferSize = 1024 * 64
	if os.Getenv("MATCH_PATTERNS") != "" {
		matchPatterns = strings.Split(os.Getenv("MATCH_PATTERNS"), ",")
	}
	if os.Getenv("FILTER_PATTERNS") != "" {
		filterPatterns = strings.Split(os.Getenv("FILTER_PATTERNS"), ",")
	}
	if os.Getenv("BUFFER_SIZE") != "" {
		size, err := strconv.Atoi(os.Getenv("BUFFER_SIZE"))
		if err != nil {
			logrus.WithField("buffer_size", os.Getenv("BUFFER_SIZE")).Error("could not parse BUFFER_SIZE env variable into an int, defaulting to 64KiB")
		} else {
			bufferSize = uint(size)
		}
	}
	if strings.ToLower(os.Getenv("FORCE_GUNZIP")) == "true" {
		forceGunzip = true
	}

	if os.Getenv("RENAME_FIELDS") != "" {
		renameFieldsConfig := strings.Split(os.Getenv("RENAME_FIELDS"), ",")
		for _, kv := range renameFieldsConfig {
			kvPair := strings.Split(kv, "=")

			if len(kvPair) != 2 {
				logrus.WithField("arg", kv).
					Error("Invalid RENAME_FIELD entry. Should be format 'before=after' ")
				continue
			}

			renameFields[kvPair[0]] = kvPair[1]
		}
	}

	lambda.Start(Handler)
}

func getReaderForKey(svc *s3.S3, bucket, key string) (io.ReadCloser, error) {
	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"bucket": bucket,
			"key":    key,
		}).Warn("unable to get object from bucket")
		return nil, err
	}

	reader := resp.Body
	// See https://github.com/aws/aws-sdk-go/issues/1292
	// The default HTTP transport that the AWS SDK uses will decompress objects transparently
	// if the Content Encoding is gzip. Not everyone or everything properly sets the Content-Encoding
	// header on their S3 objects, so we could be trying to process gzipped objects and not know it.
	// Unfortunately, to work around this, the end-user will need to tell us that's what's happening with the FORCE_GZIP env variable.

	// What if there is a mix of gzipped objects and non-gzipped objects? The only way to know is
	// to attempt to uncompress the object and see if it's gzipped. Unfortunately, this causes us to eat part of
	// the object Body, so if we're wrong, we need to retry.
	if forceGunzip {
		reader, err = gzip.NewReader(resp.Body)
		if err == nil {
			return reader, nil
		} else if err == gzip.ErrHeader {
			logrus.WithError(err).WithField("key", key).
				Warn("not a gzipped object, retrying without gzip")
		} else {
			logrus.WithError(err).WithField("key", key).
				Warn("unable to create gzip reader for object")
			return nil, err
		}
		// clean up resources - this body no good now that we've called Read
		// (we could optimize a way around this but Not Today)
		resp.Body.Close()

		// Retry fetching the object
		resp, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		})
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"bucket": bucket,
				"key":    key,
			}).Warn("unable to get object from bucket")
			return nil, err
		}
		reader = resp.Body
	}

	return reader, nil
}
