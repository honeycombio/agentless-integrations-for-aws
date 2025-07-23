#!/bin/bash

set -e

echo "=== AWS Regions List Updater ==="
echo "This script updates aws_regions_list.txt with regions where:"
echo "  1. Your AWS account is opted-in"
echo "  2. The honeycomb-integrations-{region} S3 bucket exists"
echo ""

OUTPUT_FILE="aws_regions_list.txt"
TEMP_FILE=$(mktemp)

# Get current date for the header
CURRENT_DATE=$(date +"%Y-%m-%d")

# Write header to temp file
cat > "${TEMP_FILE}" << EOF
# AWS Regions List for Honeycomb Integrations
# Generated on: ${CURRENT_DATE}
#
# This file contains AWS regions where:
#   1. This AWS account is opted-in
#   2. The honeycomb-integrations-{region} S3 bucket exists
#
# To update this list, run: ./update_aws_regions_list.sh
#

EOF

echo "Fetching enabled AWS regions (excluding China and GovCloud)..."
ENABLED_REGIONS=($(aws ec2 describe-regions \
  --query 'Regions[].RegionName' \
  --output json | \
  jq -r '.[]' | \
  grep -v '^cn-' | \
  grep -v '^us-gov-' | \
  sort))

echo "Found ${#ENABLED_REGIONS[@]} enabled regions"
echo ""

echo "Checking for S3 buckets in each region..."
echo "================================================"

REGIONS_WITH_BUCKETS=()
REGIONS_WITHOUT_BUCKETS=()

for region in "${ENABLED_REGIONS[@]}"; do
  bucket_name="honeycomb-integrations-${region}"
  printf "Checking %-20s: " "${region}"

  if aws s3api head-bucket --bucket "${bucket_name}" 2>/dev/null; then
    echo "✅ Bucket exists"
    REGIONS_WITH_BUCKETS+=("${region}")
    echo "${region}" >> "${TEMP_FILE}"
  else
    echo "❌ No bucket"
    REGIONS_WITHOUT_BUCKETS+=("${region}")
  fi
done

echo ""
echo "================================================"
echo "Summary:"
echo "  - Regions with buckets:    ${#REGIONS_WITH_BUCKETS[@]}"
echo "  - Regions without buckets: ${#REGIONS_WITHOUT_BUCKETS[@]}"
echo ""

# Move temp file to final location
mv "${TEMP_FILE}" "${OUTPUT_FILE}"
echo "✅ Updated ${OUTPUT_FILE} with ${#REGIONS_WITH_BUCKETS[@]} regions"

if [[ ${#REGIONS_WITHOUT_BUCKETS[@]} -gt 0 ]]; then
  echo ""
  echo "⚠️  The following ${#REGIONS_WITHOUT_BUCKETS[@]} regions are enabled but don't have buckets:"
  printf '%s\n' "${REGIONS_WITHOUT_BUCKETS[@]}" | column

  echo ""
  echo "To create a bucket in a region, run:"
  echo ""
  echo "  REGION=<region-name>"
  echo "  aws s3api create-bucket \\"
  echo "    --bucket honeycomb-integrations-\${REGION} \\"
  echo "    --region \${REGION} \\"
  echo "    --create-bucket-configuration LocationConstraint=\${REGION} \\"
  echo "    --acl public-read"
  echo ""
  echo "Note: For us-east-1, omit the --create-bucket-configuration parameter"
fi
