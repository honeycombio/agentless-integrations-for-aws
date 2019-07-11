package main

import (
	"bufio"
	"compress/gzip"
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
		resp, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: &record.S3.Bucket.Name,
			Key:    &record.S3.Object.Key,
		})
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"key":    record.S3.Object.Key,
				"bucket": record.S3.Bucket.Name,
			}).Warn("unable to get object from bucket")
			continue
		}

		reader := resp.Body
		// figure out if this file is gzipped
		if resp.ContentEncoding != nil && *resp.ContentEncoding == "gzip" {
			// Skip attempting to unzip, the default http client already unzipped.
			// see https://github.com/aws/aws-sdk-go/issues/1292
		} else if resp.ContentType != nil {
			if *resp.ContentType == "application/x-gzip" || *resp.ContentType == "application/octet-stream" {
				reader, err = gzip.NewReader(resp.Body)
				if err != nil {
					logrus.WithError(err).WithField("key", record.S3.Object.Key).
						Warn("unable to create gzip reader for object")
					continue
				}
			}
		} else if strings.HasSuffix(record.S3.Object.Key, ".gz") {
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				logrus.WithError(err).WithField("key", record.S3.Object.Key).
					Warn("unable to create gzip reader for object")
				continue
			}
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

	lambda.Start(Handler)
}
