#!/bin/bash
# ships handlers to S3 for use in templates

set -e

# A fallback version when run outside of CI. Will look like: 2.2.1-13-gd591456
# 2.2.1 - most recent tag with the leading v trimmed
# 13 - no.of commits away from tag
# gd591456 - this commit's id
GIT_VERSION="`git describe | sed -e s/^v//`"
VERSION="${CIRCLE_TAG:-$GIT_VERSION}"
REGIONS="us-east-1 us-east-2 us-west-1 us-west-2 ap-south-1 ap-northeast-2 ap-southeast-1 ap-southeast-2 ap-northeast-1 ca-central-1 eu-central-1 eu-west-1 eu-west-2 eu-west-3 sa-east-1"

ZIP_NAME="ingest-handlers.zip"
ZIP_PATH="./pkg/${ZIP_NAME}"

if [[ -f "${ZIP_PATH}" ]]; then
  echo "+++ Publishing ${ZIP_PATH} to S3"
else
  echo 1>&2 "$ZIP_PATH does not exist. Run build.sh?"
	exit 1
fi

# if DRYRUN is set to anything, turn it into the awscli switch
[[ -n "${DRYRUN}" ]] && DRYRUN="--dryrun"

echo "+++ Uploading handlers"
for REGION in ${REGIONS}; do
	DEPLOY_ROOT=s3://honeycomb-integrations-${REGION}/agentless-integrations-for-aws
	aws s3 cp ${DRYRUN} ${ZIP_PATH} ${DEPLOY_ROOT}/LATEST/${ZIP_NAME}
	aws s3 cp ${DRYRUN} ${ZIP_PATH} ${DEPLOY_ROOT}/${VERSION}/${ZIP_NAME}
done

# publish the templates to our builds bucket
DEPLOY_ROOT=s3://honeycomb-builds/honeycombio/integrations-for-aws

echo "+++ Uploading templates"
for TEMPLATE in templates/*; do
	aws s3 cp ${DRYRUN} ${TEMPLATE} ${DEPLOY_ROOT}/LATEST/${TEMPLATE}
	aws s3 cp ${DRYRUN} ${TEMPLATE} ${DEPLOY_ROOT}/${VERSION}/${TEMPLATE}
done
