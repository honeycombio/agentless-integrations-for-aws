#!/bin/bash
# packages handlers and ships them to S3 for use in templates

ROOT_DIR=$(pwd)
rm -rf pkg
mkdir pkg

for HANDLER in "cloudwatch-handler"; do
	cd ${HANDLER}
	GOOS=linux go build
	cd ${ROOT_DIR}
	mv ${HANDLER}/${HANDLER} pkg
done;

cd ./pkg

zip ingest-handlers.zip *

aws s3 cp ingest-handlers.zip s3://honeycomb-builds/honeycombio/serverless-ingest-poc/ingest-handlers-LATEST.zip
