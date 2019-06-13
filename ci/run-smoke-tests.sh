#!/bin/bash

set -eux

cf login -a $CF_API_URL -u $CF_USERNAME -p $CF_PASSWORD -o $CF_ORGANIZATION -s $CF_SPACE

# Clean up existing app and service if present
cf delete -f smoke-tests-$SERVICE_PLAN
cf delete-service -f rds-smoke-tests-$SERVICE_PLAN

# Create service
if echo "$SERVICE_PLAN" | grep -v shared | grep mysql >/dev/null ; then
  # test out the enable_functions stuff, which only works on non-shared mysql databases
  cf create-service aws-rds "$SERVICE_PLAN" "rds-smoke-tests-$SERVICE_PLAN" -c '{"enable_functions": true}'

  # Write manifest with ENABLE_FUNCTIONS set so that we can test for it
  cat << EOF > aws-broker-app/ci/smoke-tests/manifest.yml
---
applications:
- name: smoke-tests-${SERVICE_PLAN}
  buildpack: binary_buildpack
  command: ./smoke-tests.sh
  env:
    DB_TYPE: ${DB_TYPE}
    ENABLE_FUNCTIONS: true
  services:
  - rds-smoke-tests-${SERVICE_PLAN}
EOF
else
  # create a regular instance
  cf create-service aws-rds "$SERVICE_PLAN" "rds-smoke-tests-$SERVICE_PLAN"

  # Write manifest
  cat << EOF > aws-broker-app/ci/smoke-tests/manifest.yml
---
applications:
- name: smoke-tests-${SERVICE_PLAN}
  buildpack: binary_buildpack
  command: ./smoke-tests.sh
  env:
    DB_TYPE: ${DB_TYPE}
  services:
  - rds-smoke-tests-${SERVICE_PLAN}
EOF
fi



cp -R sqlclient-oracle-basiclite aws-broker-app/ci/smoke-tests/.
cp -R sqlclient-oracle-sqlplus aws-broker-app/ci/smoke-tests/.
cp -R sqlclient-postgres aws-broker-app/ci/smoke-tests/.
cp -R sqlclient-mysql aws-broker-app/ci/smoke-tests/.

# Wait until service is available
while true; do
  if out=$(cf push -f aws-broker-app/ci/smoke-tests/manifest.yml -p aws-broker-app/ci/smoke-tests 2>&1); then
    break
  fi
  if ! [[ $out =~ "Instance not available yet" ]]; then
    echo "${out}"
    exit 1
  fi
  sleep 90
done

# Clean up app and service
cf delete -f smoke-tests-$SERVICE_PLAN
cf delete-service -f rds-smoke-tests-$SERVICE_PLAN
