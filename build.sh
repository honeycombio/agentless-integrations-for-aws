#!/bin/bash
# packages handlers and ships them to S3 for use in templates

set -e

./test.sh

VERSION=1.0.0
DEPLOY_ROOT=s3://honeycomb-builds/honeycombio/integrations-for-aws

ROOT_DIR=$(pwd)
rm -rf pkg
mkdir pkg

HANDLERS="cloudwatch-handler s3-handler sns-handler"

for HANDLER in ${HANDLERS}; do
	cd ${HANDLER}
	GOOS=linux go build
	cd ${ROOT_DIR}
	mv ${HANDLER}/${HANDLER} pkg
done;

cd ./pkg

zip ingest-handlers.zip *

aws s3 cp ingest-handlers.zip ${DEPLOY_ROOT}/LATEST/ingest-handlers.zip
aws s3 cp ingest-handlers.zip ${DEPLOY_ROOT}/${VERSION}/ingest-handlers.zip

cd ${ROOT_DIR}

for TEMPLATE in templates/*; do
	aws s3 cp ${TEMPLATE} ${DEPLOY_ROOT}/LATEST/${TEMPLATE}
	aws s3 cp ${TEMPLATE} ${DEPLOY_ROOT}/${VERSION}/${TEMPLATE}
done
