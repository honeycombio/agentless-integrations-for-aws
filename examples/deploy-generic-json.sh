#!/bin/bash
# This script creates a Honeycomb lambda function that listens to a Cloudwatch
# Log Group receiving JSON-formatted logs and sends them as Honeycomb events.

ENVIRONMENT=production
STACK_NAME=${ENVIRONMENT}-json-logs
# change this to the log group name used by your VPC flow logs
LOG_GROUP_NAME=/${ENVIRONMENT}-json-logs
# this is the base64-encoded KMS encrypted CiphertextBlob containing your write key
# To encrypt your key, run `aws kms encrypt --key-id $MY_KMS_KEY_ID --plaintext "$MY_HONEYCOMB_KEY"`
# paste the CyphertextBlob here
HONEYCOMB_WRITE_KEY=CHANGEME
# this is the KMS Key ID used to encrypt the write key above
# try running `aws kms list-keys` - you want the UID after ":key/" in the ARN
KMS_KEY_ID=CHANGEME
DATASET="cloudwatch-logs"
HONEYCOMB_SAMPLE_RATE="1"
TEMPLATE="file://../templates/cloudwatch-logs-json.yml"

JSON=$(cat << END
{
    "StackName": "${STACK_NAME}",
    "Parameters": [
        {
            "ParameterKey": "Environment",
            "ParameterValue": "${ENVIRONMENT}"
        },
        {
            "ParameterKey": "HoneycombWriteKey",
            "ParameterValue": "${HONEYCOMB_WRITE_KEY}"
        },
        {
            "ParameterKey": "KMSKeyId",
            "ParameterValue": "${KMS_KEY_ID}"
        },
        {
            "ParameterKey": "HoneycombDataset",
            "ParameterValue": "${DATASET}"
        },
        {
            "ParameterKey": "HoneycombSampleRate",
            "ParameterValue": "${HONEYCOMB_SAMPLE_RATE}"
        },
        {
            "ParameterKey": "LogGroupName",
            "ParameterValue": "${LOG_GROUP_NAME}"
        },
        {
            "ParameterKey": "TimeFieldName",
            "ParameterValue": ""
        },
        {
            "ParameterKey": "TimeFieldFormat",
            "ParameterValue": ""
        }
    ],
    "Capabilities": [
        "CAPABILITY_IAM"
    ],
    "OnFailure": "ROLLBACK",
    "Tags": [
        {
            "Key": "Environment",
            "Value": "${ENVIRONMENT}"
        }
    ]
}
END
)

aws cloudformation create-stack --cli-input-json "${JSON}" --template-body=${TEMPLATE}
