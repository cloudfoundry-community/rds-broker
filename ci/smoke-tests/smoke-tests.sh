#!/bin/bash

set -e -x

curl -O -L https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64
JQ=./jq-linux64
chmod +x $JQ

if [ $DB_TYPE = "postgres" ] ; then
  curl https://s3.amazonaws.com/18f-cf-cli/psql-9.4.4-ubuntu-14.04.tar.gz | tar xvz
  ./psql/bin/psql $DATABASE_URL -c "create table smoke (id integer, name text);"
  ./psql/bin/psql $DATABASE_URL -c "insert into smoke values (1, 'smoke');"
  ./psql/bin/psql $DATABASE_URL -c "drop table smoke;"
elif [ $DB_TYPE = "mysql" ] ; then
  curl https://s3.amazonaws.com/18f-cf-cli/mysql.gz | gunzip > mysql
  chmod +x ./mysql
  MYSQL_HOST=`echo $VCAP_SERVICES | $JQ -c -r '.["aws-rds"] | .[0].credentials.host'`
  MYSQL_USER=`echo $VCAP_SERVICES | $JQ -c -r '.["aws-rds"] | .[0].credentials.username'`
  MYSQL_PASS=`echo $VCAP_SERVICES | $JQ -c -r '.["aws-rds"] | .[0].credentials.password'`
  MYSQL_DB=`echo $VCAP_SERVICES | $JQ -c -r '.["aws-rds"] | .[0].credentials.db_name'`
  ./mysql -h $MYSQL_HOST -u $MYSQL_USER -p$MYSQL_PASS -e "create table smoke (id integer, name text);" $MYSQL_DB
  ./mysql -h $MYSQL_HOST -u $MYSQL_USER -p$MYSQL_PASS -e "insert into smoke values (1, 'smoke');" $MYSQL_DB
  ./mysql -h $MYSQL_HOST -u $MYSQL_USER -p$MYSQL_PASS -e "drop table smoke;" $MYSQL_DB
else
  echo "\$DB_TYPE must be postgres or mysql"
  exit 1
fi

sleep infinity
