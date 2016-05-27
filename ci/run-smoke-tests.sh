#!/bin/bash

set -e

cf login -a $CF_API_URL -u $CF_DEPLOY_USERNAME -p $CF_DEPLOY_PASSWORD -o $CF_ORGANIZATION -s $CF_SPACE

cf create-service rds $SERVICE_PLAN rds-smoke-test

while true; do
  if cf push -f aws-broker-app/ci/smoke-tests/manifest.yml -p aws-broker-app/ci/smoke-tests > out.txt ; then
    break
  fi
  if ! grep "Instance not available yet" out.txt ; then
    cat out.txt
    exit 1
  fi
  sleep 10
done

cf delete -f smoke-tests
cf delete-service -f rds-smoke-test
