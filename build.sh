#!/bin/bash
# packages handlers and ships them to S3 for use in templates

set -e

./test.sh

VERSION=1.10.0
REGIONS="us-east-1 us-east-2 us-west-1 us-west-2 ap-south-1 ap-northeast-2 ap-southeast-1 ap-southeast-2 ap-northeast-1 ca-central-1 eu-central-1 eu-west-1 eu-west-2 eu-west-3 sa-east-1"

ROOT_DIR=$(pwd)
rm -rf pkg
mkdir pkg

HANDLERS="cloudwatch-handler s3-handler sns-handler mysql-handler postgresql-handler publisher"

for HANDLER in ${HANDLERS}; do
	cd ${HANDLER}
	GOOS=linux go build
	cd ${ROOT_DIR}
	mv ${HANDLER}/${HANDLER} pkg
done;

cd ./pkg

zip ingest-handlers.zip *

for REGION in ${REGIONS}; do
	DEPLOY_ROOT=s3://honeycomb-integrations-${REGION}/agentless-integrations-for-aws
	aws s3 cp ingest-handlers.zip ${DEPLOY_ROOT}/LATEST/ingest-handlers.zip
	aws s3 cp ingest-handlers.zip ${DEPLOY_ROOT}/${VERSION}/ingest-handlers.zip
done;

cd ${ROOT_DIR}

# publish the templates to our builds bucket
DEPLOY_ROOT=s3://honeycomb-builds/honeycombio/integrations-for-aws

for TEMPLATE in templates/*; do
	aws s3 cp ${TEMPLATE} ${DEPLOY_ROOT}/LATEST/${TEMPLATE}
	aws s3 cp ${TEMPLATE} ${DEPLOY_ROOT}/${VERSION}/${TEMPLATE}
done
