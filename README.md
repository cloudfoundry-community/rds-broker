## Cloud Foundry RDS Service Broker

Cloud Foundry Service Broker to manage RDS instances and a shared RDS Database.

### How to deploy it

1. `cf push`
1. `cf create-service-broker SERVICE-NAME USER PASS https://BROKER-URL/`
1. `cf enable-service-access rds-database`


### How to use it

To use the service you need to create a service instance and bind it:

1. `cf create-service rds-database shared-psql MYDB`
1. `cf bind-service APP MYDB`

When you do that you will have all the credentials in the 
`VCAP_SERVICES` environment variable with the JSON key `rds-database`.

Also, you will have a `DATABASE_URL` environment variable that will
be the connection string to the DB.
