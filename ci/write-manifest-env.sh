#!/bin/sh

set -e -x

cat $STATE_FILE | \
  jq '.modules[1]["resources"]["aws_db_instance.rds_database"]["primary"]["attributes"]' > \
  attributes.json

cat << EOF > $OUTPUT_FILE
env:
  DB_PORT: `cat attributes.json | jq '.port'`
  DB_NAME: `cat attributes.json | jq '.name'`
  DB_USER: `cat attributes.json | jq '.username'`
  DB_PASS: `cat attributes.json | jq '.password'`
  DB_TYPE: `cat attributes.json | jq '.engine'`
  DB_URL: `cat attributes.json | jq '.endpoint'`
EOF
