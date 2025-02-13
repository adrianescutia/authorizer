package sql

import (
	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/db/models"
	"github.com/authorizerdev/authorizer/server/envstore"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type provider struct {
	db *gorm.DB
}

// NewProvider returns a new SQL provider
func NewProvider() (*provider, error) {
	var sqlDB *gorm.DB
	var err error

	ormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: models.Prefix,
		},
	}

	switch envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyDatabaseType) {
	case constants.DbTypePostgres:
		sqlDB, err = gorm.Open(postgres.Open(envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyDatabaseURL)), ormConfig)
		break
	case constants.DbTypeSqlite:
		sqlDB, err = gorm.Open(sqlite.Open(envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyDatabaseURL)), ormConfig)
		break
	case constants.DbTypeMysql:
		sqlDB, err = gorm.Open(mysql.Open(envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyDatabaseURL)), ormConfig)
		break
	case constants.DbTypeSqlserver:
		sqlDB, err = gorm.Open(sqlserver.Open(envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyDatabaseURL)), ormConfig)
		break
	}

	if err != nil {
		return nil, err
	}

	sqlDB.AutoMigrate(&models.User{}, &models.VerificationRequest{}, &models.Session{}, &models.Env{})
	return &provider{
		db: sqlDB,
	}, nil
}
