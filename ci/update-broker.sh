#!/bin/bash

set -e

cf login -a $CF_API_URL -u $CF_DEPLOY_USERNAME -p $CF_DEPLOY_PASSWORD -o $CF_ORGANIZATION -s $CF_SPACE

BROKER_URL=https://`cf app $BROKER_NAME | grep urls: | sed 's/urls: //'`

# Create or update service broker
if ! cf create-service-broker $BROKER_NAME $AUTH_USER $AUTH_PASS $BROKER_URL ; then
  cf update-service-broker $BROKER_NAME $AUTH_USER $AUTH_PASS $BROKER_URL
fi

# Allow broker access to all orgs
cf enable-service-access $PLAN_NAME
