package resolvers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/db"
	"github.com/authorizerdev/authorizer/server/db/models"
	"github.com/authorizerdev/authorizer/server/email"
	"github.com/authorizerdev/authorizer/server/envstore"
	"github.com/authorizerdev/authorizer/server/graph/model"
	"github.com/authorizerdev/authorizer/server/token"
	"github.com/authorizerdev/authorizer/server/utils"
)

// MagicLinkLoginResolver is a resolver for magic link login mutation
func MagicLinkLoginResolver(ctx context.Context, params model.MagicLinkLoginInput) (*model.Response, error) {
	var res *model.Response

	if envstore.EnvInMemoryStoreObj.GetBoolStoreEnvVariable(constants.EnvKeyDisableMagicLinkLogin) {
		return res, fmt.Errorf(`magic link login is disabled for this instance`)
	}

	params.Email = strings.ToLower(params.Email)

	if !utils.IsValidEmail(params.Email) {
		return res, fmt.Errorf(`invalid email address`)
	}

	inputRoles := []string{}

	user := models.User{
		Email: params.Email,
	}

	// find user with email
	existingUser, err := db.Provider.GetUserByEmail(params.Email)

	if err != nil {
		user.SignupMethods = constants.SignupMethodMagicLinkLogin
		// define roles for new user
		if len(params.Roles) > 0 {
			// check if roles exists
			if !utils.IsValidRoles(envstore.EnvInMemoryStoreObj.GetSliceStoreEnvVariable(constants.EnvKeyRoles), params.Roles) {
				return res, fmt.Errorf(`invalid roles`)
			} else {
				inputRoles = params.Roles
			}
		} else {
			inputRoles = envstore.EnvInMemoryStoreObj.GetSliceStoreEnvVariable(constants.EnvKeyDefaultRoles)
		}

		user.Roles = strings.Join(inputRoles, ",")
		user, _ = db.Provider.AddUser(user)
	} else {
		user = existingUser
		// There multiple scenarios with roles here in magic link login
		// 1. user has access to protected roles + roles and trying to login
		// 2. user has not signed up for one of the available role but trying to signup.
		// 		Need to modify roles in this case

		// find the unassigned roles
		existingRoles := strings.Split(existingUser.Roles, ",")
		unasignedRoles := []string{}
		for _, ir := range inputRoles {
			if !utils.StringSliceContains(existingRoles, ir) {
				unasignedRoles = append(unasignedRoles, ir)
			}
		}

		if len(unasignedRoles) > 0 {
			// check if it contains protected unassigned role
			hasProtectedRole := false
			for _, ur := range unasignedRoles {
				if utils.StringSliceContains(envstore.EnvInMemoryStoreObj.GetSliceStoreEnvVariable(constants.EnvKeyProtectedRoles), ur) {
					hasProtectedRole = true
				}
			}

			if hasProtectedRole {
				return res, fmt.Errorf(`invalid roles`)
			} else {
				user.Roles = existingUser.Roles + "," + strings.Join(unasignedRoles, ",")
			}
		} else {
			user.Roles = existingUser.Roles
		}

		signupMethod := existingUser.SignupMethods
		if !strings.Contains(signupMethod, constants.SignupMethodMagicLinkLogin) {
			signupMethod = signupMethod + "," + constants.SignupMethodMagicLinkLogin
		}

		user.SignupMethods = signupMethod
		user, _ = db.Provider.UpdateUser(user)
		if err != nil {
			log.Println("error updating user:", err)
		}
	}

	if !envstore.EnvInMemoryStoreObj.GetBoolStoreEnvVariable(constants.EnvKeyDisableEmailVerification) {
		// insert verification request
		verificationType := constants.VerificationTypeMagicLinkLogin
		verificationToken, err := token.CreateVerificationToken(params.Email, verificationType)
		if err != nil {
			log.Println(`error generating token`, err)
		}
		db.Provider.AddVerificationRequest(models.VerificationRequest{
			Token:      verificationToken,
			Identifier: verificationType,
			ExpiresAt:  time.Now().Add(time.Minute * 30).Unix(),
			Email:      params.Email,
		})

		// exec it as go routin so that we can reduce the api latency
		go func() {
			email.SendVerificationMail(params.Email, verificationToken)
		}()
	}

	res = &model.Response{
		Message: `Magic Link has been sent to your email. Please check your inbox!`,
	}

	return res, nil
}
