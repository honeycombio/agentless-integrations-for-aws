#!/bin/bash
# Very simple build script that only needs the AWSCLI to deploy.

# This is the role that the function will run under. It should
# already exist
ROLE=arn:aws:iam::702835727665:role/lambda_basic_execution
NAME=honeycomb-cloudwatch-integration
REGION=us-east-1

ROOT_DIR=$(pwd)
rm -rf pkg
mkdir pkg

for HANDLER in "cloudwatch-handler"; do
	cd ${HANDLER}
	GOOS=linux go build
	cd ${ROOT_DIR}
	cp ${HANDLER}/${HANDLER} pkg
done;

cd ./pkg

zip ingest-handlers.zip *

aws s3 cp ingest-handlers.zip s3://honeycomb-builds/honeycombio/serverless-ingest-poc/ingest-handlers-LATEST.zip
