#!/bin/bash
# quickly validate that templates are valid

set -e
export AWS_PAGER=""
for TEMPLATE in templates/*.yml; do
    aws cloudformation validate-template --template-body=file://./${TEMPLATE} \
    && echo "${TEMPLATE} parses OK."
done;
