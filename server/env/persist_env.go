package env

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/db"
	"github.com/authorizerdev/authorizer/server/db/models"
	"github.com/authorizerdev/authorizer/server/envstore"
	"github.com/authorizerdev/authorizer/server/utils"
	"github.com/google/uuid"
)

// PersistEnv persists the environment variables to the database
func PersistEnv() error {
	env, err := db.Provider.GetEnv()
	// config not found in db
	if err != nil {
		// AES encryption needs 32 bit key only, so we chop off last 4 characters from 36 bit uuid
		hash := uuid.New().String()[:36-4]
		envstore.EnvInMemoryStoreObj.UpdateEnvVariable(constants.StringStoreIdentifier, constants.EnvKeyEncryptionKey, hash)
		encodedHash := utils.EncryptB64(hash)

		configData, err := json.Marshal(envstore.EnvInMemoryStoreObj.GetEnvStoreClone())
		if err != nil {
			return err
		}

		encryptedConfig, err := utils.EncryptAES(configData)
		if err != nil {
			return err
		}

		env = models.Env{
			Hash:    encodedHash,
			EnvData: encryptedConfig,
		}

		db.Provider.AddEnv(env)
	} else {
		// decrypt the config data from db
		// decryption can be done using the hash stored in db
		encryptionKey := env.Hash
		decryptedEncryptionKey, err := utils.DecryptB64(encryptionKey)
		if err != nil {
			return err
		}

		envstore.EnvInMemoryStoreObj.UpdateEnvVariable(constants.StringStoreIdentifier, constants.EnvKeyEncryptionKey, decryptedEncryptionKey)
		decryptedConfigs, err := utils.DecryptAES(env.EnvData)
		if err != nil {
			return err
		}

		// temp store variable
		var storeData envstore.Store

		err = json.Unmarshal(decryptedConfigs, &storeData)
		if err != nil {
			return err
		}

		// if env is changed via env file or OS env
		// give that higher preference and update db, but we don't recommend it

		hasChanged := false

		for key, value := range storeData.StringEnv {
			if key != constants.EnvKeyEncryptionKey {
				// check only for derivative keys
				// No need to check for ENCRYPTION_KEY which special key we use for encrypting config data
				// as we have removed it from json
				envValue := strings.TrimSpace(os.Getenv(key))

				// env is not empty
				if envValue != "" {
					if value != envValue {
						storeData.StringEnv[key] = envValue
						hasChanged = true
					}
				}
			}
		}

		for key, value := range storeData.BoolEnv {
			envValue := strings.TrimSpace(os.Getenv(key))
			// env is not empty
			if envValue != "" {
				envValueBool, _ := strconv.ParseBool(envValue)
				if value != envValueBool {
					storeData.BoolEnv[key] = envValueBool
					hasChanged = true
				}
			}
		}

		for key, value := range storeData.SliceEnv {
			envValue := strings.TrimSpace(os.Getenv(key))
			// env is not empty
			if envValue != "" {
				envStringArr := strings.Split(envValue, ",")
				if !utils.IsStringArrayEqual(value, envStringArr) {
					storeData.SliceEnv[key] = envStringArr
					hasChanged = true
				}
			}
		}

		// handle derivative cases like disabling email verification & magic login
		// in case SMTP is off but env is set to true
		if storeData.StringEnv[constants.EnvKeySmtpHost] == "" || storeData.StringEnv[constants.EnvKeySmtpUsername] == "" || storeData.StringEnv[constants.EnvKeySmtpPassword] == "" || storeData.StringEnv[constants.EnvKeySenderEmail] == "" && storeData.StringEnv[constants.EnvKeySmtpPort] == "" {
			if !storeData.BoolEnv[constants.EnvKeyDisableEmailVerification] {
				storeData.BoolEnv[constants.EnvKeyDisableEmailVerification] = true
				hasChanged = true
			}

			if !storeData.BoolEnv[constants.EnvKeyDisableMagicLinkLogin] {
				storeData.BoolEnv[constants.EnvKeyDisableMagicLinkLogin] = true
				hasChanged = true
			}
		}

		envstore.EnvInMemoryStoreObj.UpdateEnvStore(storeData)
		if hasChanged {
			encryptedConfig, err := utils.EncryptEnvData(storeData)
			if err != nil {
				return err
			}

			env.EnvData = encryptedConfig
			_, err = db.Provider.UpdateEnv(env)
			if err != nil {
				log.Println("error updating config:", err)
				return err
			}
		}

	}

	return nil
}
