#!/bin/bash

set -e

# A fallback version when run outside of CI. Will look like: 2.2.1-13-gd591456
# 2.2.1 - most recent tag with the leading v trimmed
# 13 - no.of commits away from tag
# gd591456 - this commit's id
GIT_VERSION="`git describe | sed -e s/^v//`"
VERSION="${CIRCLE_TAG:-$GIT_VERSION}"
# Read AWS regions from file maintained by update_aws_regions_list.sh
REGIONS_FILE="aws_regions_list.txt"

if [[ ! -f "${REGIONS_FILE}" ]]; then
  echo "ERROR: ${REGIONS_FILE} not found. Run ./update_aws_regions_list.sh to generate it."
  exit 1
fi

echo "+++ Reading AWS regions from ${REGIONS_FILE}"
REGIONS=()
while IFS= read -r region; do
  # Skip empty lines and comments
  [[ -z "$region" || "$region" == \#* ]] && continue
  REGIONS+=("$region")
done < "${REGIONS_FILE}"

echo "+++ Will deploy to ${#REGIONS[@]} regions: ${REGIONS[*]}"
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

echo "+++ Uploading handlers to ${#REGIONS[@]} regions"
for HANDLER in "${HANDLERS[@]}"; do
  for REGION in "${REGIONS[@]}"; do
    DEPLOY_ROOT=s3://honeycomb-integrations-${REGION}/agentless-integrations-for-aws
    echo "Deploying ${HANDLER} to ${REGION}..."

    aws s3 cp ${DRYRUN} ${ZIP_PATH}/${HANDLER}-amd64.zip ${DEPLOY_ROOT}/${VERSION}/${HANDLER}-amd64.zip
    [[ -n "$CIRCLE_TAG" ]] && aws s3 cp ${DRYRUN} ${ZIP_PATH}/${HANDLER}-amd64.zip ${DEPLOY_ROOT}/LATEST/${HANDLER}-amd64.zip || true
    aws s3 cp ${DRYRUN} ${ZIP_PATH}/${HANDLER}-arm64.zip ${DEPLOY_ROOT}/${VERSION}/${HANDLER}-arm64.zip
    [[ -n "$CIRCLE_TAG" ]] && aws s3 cp ${DRYRUN} ${ZIP_PATH}/${HANDLER}-arm64.zip ${DEPLOY_ROOT}/LATEST/${HANDLER}-arm64.zip || true
  done
done

echo "+++ Successfully deployed all handlers to ${#REGIONS[@]} regions"
