package main

import (
	"flag"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/db"
	"github.com/authorizerdev/authorizer/server/env"
	"github.com/authorizerdev/authorizer/server/envstore"
	"github.com/authorizerdev/authorizer/server/oauth"
	"github.com/authorizerdev/authorizer/server/routes"
	"github.com/authorizerdev/authorizer/server/sessionstore"
)

var VERSION string

func main() {
	envstore.ARG_DB_URL = flag.String("database_url", "", "Database connection string")
	envstore.ARG_DB_TYPE = flag.String("database_type", "", "Database type, possible values are postgres,mysql,sqlite")
	envstore.ARG_ENV_FILE = flag.String("env_file", "", "Env file path")
	flag.Parse()

	envstore.EnvInMemoryStoreObj.UpdateEnvVariable(constants.StringStoreIdentifier, constants.EnvKeyVersion, VERSION)

	env.InitEnv()
	db.InitDB()
	env.PersistEnv()

	sessionstore.InitSession()
	oauth.InitOAuth()

	router := routes.InitRouter()

	router.Run(":" + envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyPort))
}
