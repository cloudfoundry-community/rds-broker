#!/bin/bash

set -e -x

curl -O -L https://github.com/stedolan/jq/releases/download/jq-1.5/jq-linux64
JQ=./jq-linux64
chmod +x $JQ

if [ $DB_TYPE = "postgres" ] ; then
  tar -xvzf $(find sqlclient-postgres -type f -name "psql*")
  ./psql/bin/psql $DATABASE_URL -c "create table smoke (id integer, name text);"
  ./psql/bin/psql $DATABASE_URL -c "insert into smoke values (1, 'smoke');"
  ./psql/bin/psql $DATABASE_URL -c "drop table smoke;"
elif [ $DB_TYPE = "mysql" ] ; then
  gunzip -c $(find sqlclient-mysql -type f -name "mysql*") > mysql
  chmod +x ./mysql
  MYSQL_HOST=`echo $VCAP_SERVICES | $JQ -c -r '.["aws-rds"] | .[0].credentials.host'`
  MYSQL_USER=`echo $VCAP_SERVICES | $JQ -c -r '.["aws-rds"] | .[0].credentials.username'`
  MYSQL_PASS=`echo $VCAP_SERVICES | $JQ -c -r '.["aws-rds"] | .[0].credentials.password'`
  MYSQL_DB=`echo $VCAP_SERVICES | $JQ -c -r '.["aws-rds"] | .[0].credentials.db_name'`
  ./mysql -h $MYSQL_HOST -u $MYSQL_USER -p$MYSQL_PASS -e "create table smoke (id integer, name text);" $MYSQL_DB
  ./mysql -h $MYSQL_HOST -u $MYSQL_USER -p$MYSQL_PASS -e "insert into smoke values (1, 'smoke');" $MYSQL_DB
  ./mysql -h $MYSQL_HOST -u $MYSQL_USER -p$MYSQL_PASS -e "drop table smoke;" $MYSQL_DB
  ./mysql -h $MYSQL_HOST -u $MYSQL_USER -p$MYSQL_PASS -e "create function hello(id INT) returns CHAR(50) return 'foobar';" $MYSQL_DB
elif [ $DB_TYPE = "oracle-se1" ] || [ $DB_TYPE = "oracle-se2" ] || [ $DB_TYPE = "oracle-ee" ] ; then
  unzip $(find sqlclient-oracle-basiclite -type f -name "instantclient*")
  unzip $(find sqlclient-oracle-sqlplus -type f -name "instantclient*")
  ORCL_PATH=$(find . -type d -name "instantclient*")
  SQL_HOST=$(echo $VCAP_SERVICES | $JQ -r '."aws-rds"[0].credentials.host')
  SQL_USER=$(echo $VCAP_SERVICES | $JQ -r '."aws-rds"[0].credentials.username')
  SQL_PASS=$(echo $VCAP_SERVICES | $JQ -r '."aws-rds"[0].credentials.password')
  SQL_DB=$(echo $VCAP_SERVICES | $JQ -r '."aws-rds"[0].credentials.db_name')
  # http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_ConnectToOracleInstance.html notes
  # that this connection string may fail if SQL_HOST string is > 63 chars, but
  # works OK with 80 chars in testing
  export LD_LIBRARY_PATH="/home/vcap/app/${ORCL_PATH}"
  cat <<END_SMOKE > smoke.sql
WHENEVER SQLERROR EXIT SQL.SQLCODE
CREATE TABLE smoke (id integer, name varchar2(10));
INSERT INTO smoke VALUES (1, 'smoke');
DROP TABLE smoke;
EXIT;
END_SMOKE
$ORCL_PATH/sqlplus -S "${SQL_USER}/${SQL_PASS}@${SQL_HOST}:1521/$SQL_DB" @smoke.sql
else
  echo "\$DB_TYPE must be one of: postgres mysql oracle-se1" # oracle-se2 oracle-ee
  exit 1
fi

python -m SimpleHTTPServer $PORT
