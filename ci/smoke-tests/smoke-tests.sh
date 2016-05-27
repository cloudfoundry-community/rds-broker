#!/bin/sh

set -e -x

curl https://s3.amazonaws.com/18f-cf-cli/psql-9.4.4-ubuntu-14.04.tar.gz | tar xvz

./psql/bin/psql $DATABASE_URL -c "create table smoke (id integer, name text);"
./psql/bin/psql $DATABASE_URL -c "insert into smoke values (1, 'smoke');"
./psql/bin/psql $DATABASE_URL -c "drop table smoke;"

sleep infinity
