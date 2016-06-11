#!/bin/sh

set -e -x

# Copy application to output
cp -r aws-broker-app/* built

# Fetch remote stack state
aws s3 cp s3://$S3_TFSTATE_BUCKET/$BASE_STACK_NAME/terraform.tfstate stack.tfstate

# Append environment variables to manifest
cat << EOF >> built/manifest.yml
env:
  DB_URL: `terraform output -state=$STATE_FILE -module=rds_internal rds_host`
  DB_PORT: `terraform output -state=$STATE_FILE -module=rds_internal rds_port`
EOF

# Build secrets for merging into templates
cat << EOF > built/credentials.yml
meta:
  environment: $ENVIRONMENT
  aws_broker:
    subnet_group: `terraform output -state stack.tfstate rds_subnet_group`
    postgres_security_group: `terraform output -state=stack.tfstate rds_postgres_security_group`
    mysql_security_group: `terraform output -state=stack.tfstate rds_mysql_security_group`
  shared_mysql:
    name: $RDS_SHARED_MYSQL_NAME
    username: $RDS_SHARED_MYSQL_USERNAME
    password: $RDS_SHARED_MYSQL_PASSWORD
    url: `terraform output -state=$STATE_FILE -module=rds_shared_mysql rds_host`
    port: `terraform output -state=$STATE_FILE -module=rds_shared_mysql rds_port`
  shared_postgres:
    name: $RDS_SHARED_POSTGRES_NAME
    username: $RDS_SHARED_POSTGRES_USERNAME
    password: $RDS_SHARED_POSTGRES_PASSWORD
    url: `terraform output -state=$STATE_FILE -module=rds_shared_postgres rds_host`
    port: `terraform output -state=$STATE_FILE -module=rds_shared_postgres rds_port`
EOF

# Merge secrets into templates
spiff merge aws-broker-app/secrets-template.yml built/credentials.yml > built/secrets.yml
spiff merge aws-broker-app/catalog-template.yml built/credentials.yml > built/catalog.yml
