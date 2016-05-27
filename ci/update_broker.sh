#!/bin/sh

set -e

if ! [ cf create-service-broker $BROKER_NAME $AUTH_USER $AUTH_PASS $BROKER_URL] ; then
  cf update-service-broker $BROKER_NAME $AUTH_USER $AUTH_PASS $BROKER_URL
fi

cf enable-service-access $BROKER_NAME
