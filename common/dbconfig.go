package common

// DBConfig holds configuration information to connect to a database.
// Parameters for the config.
// * dbname - The name of the database to connect to
// * user - The user to sign in as
// * password - The user's password
// * host - The host to connect to. Values that start with / are for unix domain sockets.
//   (default is localhost)
// * port - The port to bind to. (default is 5432)
// * sslmode - Whether or not to use SSL (default is require, this is not the default for libpq)
//   Valid SSL modes:
//    * disable - No SSL
//    * require - Always SSL (skip verification)
//    * verify-full - Always SSL (require verification)
type DBConfig struct {
	DbType   string
	Url      string
	Username string
	Password string
	DbName   string
	Sslmode  string
	Port     int64 // Is int64 to match the type that rds.Endpoint.Port is in the AWS RDS SDK.
}
