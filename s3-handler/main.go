package main

import (
	"bufio"
	"compress/gzip"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/serverless-ingest-poc/common"
)

// Response is a simple structured response
type Response struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

var parser parsers.LineParser
var parserType, timeFieldName, timeFieldFormat, env string

func Handler(request events.S3Event) (Response, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	config := &aws.Config{}
	svc := s3.New(sess, config)

	for _, record := range request.Records {
		resp, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: &record.S3.Bucket.Name,
			Key:    &record.S3.Object.Key,
		})
		if err != nil {
			log.Printf("unable to get object %s from bucket %s",
				record.S3.Object.Key,
				record.S3.Bucket.Name,
			)
			continue
		}

		reader := resp.Body
		// figure out if this file is gzipped
		if resp.ContentType != nil {
			if *resp.ContentType == "application/x-gzip" {
				reader, err = gzip.NewReader(resp.Body)
				if err != nil {
					log.Printf("unable to create gzip reader for %s: %s",
						record.S3.Object.Key,
						err)
					continue
				}
			}
		} else if strings.HasSuffix(record.S3.Object.Key, ".gz") {
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				log.Printf("unable to create gzip reader for %s: %s",
					record.S3.Object.Key,
					err)
				continue
			}
		}
		linesRead := 0
		scanner := bufio.NewScanner(reader)
		ok := scanner.Scan()
		for ok {
			linesRead++
			if linesRead%10000 == 0 {
				log.Printf("have processed %d lines of %s", linesRead, record.S3.Object.Key)
			}
			parsedLine, err := parser.ParseLine(scanner.Text())
			if err != nil {
				log.Printf("error parsing line: %s - line was: %s", err, scanner.Text())
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

			hnyEvent.AddField("env", env)
			hnyEvent.Send()
			ok = scanner.Scan()
		}

		if scanner.Err() != nil {
			log.Printf("s3 read of %s ended early due to err: %s",
				record.S3.Object.Key,
				err,
			)
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
		log.Fatalf("Unable to initialize libhoney with the supplied environment variables")
		return
	}
	defer libhoney.Close()

	parserType = os.Getenv("PARSER_TYPE")
	parser, err = common.ConstructParser(parserType)
	if err != nil {
		log.Fatalf("Unable to construct '%s' parser: %s", parserType, err)
		return
	}

	env = os.Getenv("ENVIRONMENT")
	timeFieldName = os.Getenv("TIME_FIELD_NAME")
	timeFieldFormat = os.Getenv("TIME_FIELD_FORMAT")

	lambda.Start(Handler)
}
