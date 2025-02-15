package resolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/db"
	"github.com/authorizerdev/authorizer/server/envstore"
	"github.com/authorizerdev/authorizer/server/graph/model"
	"github.com/authorizerdev/authorizer/server/token"
	"github.com/authorizerdev/authorizer/server/utils"
)

// ResetPasswordResolver is a resolver for reset password mutation
func ResetPasswordResolver(ctx context.Context, params model.ResetPasswordInput) (*model.Response, error) {
	var res *model.Response
	if envstore.EnvInMemoryStoreObj.GetBoolStoreEnvVariable(constants.EnvKeyDisableBasicAuthentication) {
		return res, fmt.Errorf(`basic authentication is disabled for this instance`)
	}

	verificationRequest, err := db.Provider.GetVerificationRequestByToken(params.Token)
	if err != nil {
		return res, fmt.Errorf(`invalid token`)
	}

	if params.Password != params.ConfirmPassword {
		return res, fmt.Errorf(`passwords don't match`)
	}

	// verify if token exists in db
	claim, err := token.VerifyVerificationToken(params.Token)
	if err != nil {
		return res, fmt.Errorf(`invalid token`)
	}

	user, err := db.Provider.GetUserByEmail(claim.Email)
	if err != nil {
		return res, err
	}

	password, _ := utils.EncryptPassword(params.Password)
	user.Password = &password

	signupMethod := user.SignupMethods
	if !strings.Contains(signupMethod, constants.SignupMethodBasicAuth) {
		signupMethod = signupMethod + "," + constants.SignupMethodBasicAuth
	}
	user.SignupMethods = signupMethod

	// helpful if user has not signed up with basic auth
	if user.EmailVerifiedAt == nil {
		now := time.Now().Unix()
		user.EmailVerifiedAt = &now
	}

	// delete from verification table
	db.Provider.DeleteVerificationRequest(verificationRequest)
	db.Provider.UpdateUser(user)

	res = &model.Response{
		Message: `Password updated successfully.`,
	}

	return res, nil
}
