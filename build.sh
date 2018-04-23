#!/bin/bash
# packages handlers and ships them to S3 for use in templates

set -e

# TODO: update version to incorporate Travis build number
# for now, bump this when necessary
VERSION=0.0.2

ROOT_DIR=$(pwd)
rm -rf pkg
mkdir pkg

HANDLERS="cloudwatch-handler s3-handler"

for HANDLER in ${HANDLERS}; do
	cd ${HANDLER}
	GOOS=linux go build
	cd ${ROOT_DIR}
	mv ${HANDLER}/${HANDLER} pkg
done;

cd ./pkg

zip ingest-handlers.zip *

aws s3 cp ingest-handlers.zip s3://honeycomb-builds/honeycombio/serverless-agent/LATEST/ingest-handlers.zip
aws s3 cp ingest-handlers.zip s3://honeycomb-builds/honeycombio/serverless-agent/${VERSION}/ingest-handlers.zip

cd ${ROOT_DIR}

for TEMPLATE in templates/*; do
	aws s3 cp ${TEMPLATE} s3://honeycomb-builds/honeycombio/serverless-agent/LATEST/${TEMPLATE}
	aws s3 cp ${TEMPLATE} s3://honeycomb-builds/honeycombio/serverless-agent/${VERSION}/${TEMPLATE}
done
