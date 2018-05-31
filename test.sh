#!/bin/bash
# quickly validate that templates are valid

set -e

for TEMPLATE in templates/*.yml; do
    aws cloudformation validate-template --template-body=file://./${TEMPLATE}
done;
