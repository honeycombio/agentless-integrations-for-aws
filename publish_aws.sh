#!/bin/bash

set -e

# A fallback version when run outside of CI. Will look like: 2.2.1-13-gd591456
# 2.2.1 - most recent tag with the leading v trimmed
# 13 - no.of commits away from tag
# gd591456 - this commit's id
GIT_VERSION="`git describe | sed -e s/^v//`"
VERSION="${CIRCLE_TAG:-$GIT_VERSION}"
REGIONS=(
  # US Regions
  "us-east-1"      # N. Virginia
  "us-east-2"      # Ohio
  "us-west-1"      # N. California
  "us-west-2"      # Oregon

  # Canada Regions
  "ca-central-1"   # Central
  "ca-west-1"      # Calgary

  # South America Regions
  "sa-east-1"      # São Paulo

  # Europe Regions
  "eu-central-1"   # Frankfurt
  "eu-central-2"   # Zurich
  "eu-west-1"      # Ireland
  "eu-west-2"      # London
  "eu-west-3"      # Paris
  "eu-north-1"     # Stockholm
  "eu-south-1"     # Milan
  "eu-south-2"     # Spain

  # Asia Pacific Regions
  "ap-northeast-1" # Tokyo
  "ap-northeast-2" # Seoul
  "ap-northeast-3" # Osaka
  "ap-south-1"     # Mumbai
  "ap-south-2"     # Hyderabad
  "ap-southeast-1" # Singapore
  "ap-southeast-2" # Sydney
  "ap-southeast-3" # Jakarta
  "ap-southeast-4" # Melbourne
  "ap-southeast-5" # Malaysia

  # Middle East Regions
  "me-south-1"     # Bahrain
  "me-central-1"   # UAE

  # Israel Region
  "il-central-1"   # Tel Aviv

  # Africa Region
  "af-south-1"     # Cape Town
)
HANDLERS=(
  "cloudwatch-handler"
  "s3-handler"
  "sns-handler"
  "mysql-handler"
  "postgresql-handler"
  "publisher"
  "rds-mysql-kfh-transform"
  "rds-postgresql-kfh-transform"
  "sns-kfh-transform"
)

# if DRYRUN is set to anything, turn it into the awscli switch
[[ -n "${DRYRUN}" ]] && DRYRUN="--dryrun"

ZIP_PATH="./pkg"

if [[ -d "${ZIP_PATH}" ]]; then
  echo "+++ Publishing ${ZIP_PATH} to S3"
else
  echo 1>&2 "$ZIP_PATH does not exist. Run build.sh?"
  exit 1
fi

echo "+++ Uploading handlers"
for HANDLER in "${HANDLERS[@]}"; do
  for REGION in "${REGIONS[@]}"; do
    DEPLOY_ROOT=s3://honeycomb-integrations-${REGION}/agentless-integrations-for-aws
    aws s3 cp ${DRYRUN} ${ZIP_PATH}/${HANDLER}-amd64.zip ${DEPLOY_ROOT}/${VERSION}/${HANDLER}-amd64.zip
    [[ -n "$CIRCLE_TAG" ]] && aws s3 cp ${DRYRUN} ${ZIP_PATH}/${HANDLER}-amd64.zip ${DEPLOY_ROOT}/LATEST/${HANDLER}-amd64.zip || true
    aws s3 cp ${DRYRUN} ${ZIP_PATH}/${HANDLER}-arm64.zip ${DEPLOY_ROOT}/${VERSION}/${HANDLER}-arm64.zip
    [[ -n "$CIRCLE_TAG" ]] && aws s3 cp ${DRYRUN} ${ZIP_PATH}/${HANDLER}-arm64.zip ${DEPLOY_ROOT}/LATEST/${HANDLER}-arm64.zip || true
  done
done
