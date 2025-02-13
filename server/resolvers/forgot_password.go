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

// ForgotPasswordResolver is a resolver for forgot password mutation
func ForgotPasswordResolver(ctx context.Context, params model.ForgotPasswordInput) (*model.Response, error) {
	gc, err := utils.GinContextFromContext(ctx)
	var res *model.Response
	if err != nil {
		return res, err
	}
	if envstore.EnvInMemoryStoreObj.GetBoolStoreEnvVariable(constants.EnvKeyDisableBasicAuthentication) {
		return res, fmt.Errorf(`basic authentication is disabled for this instance`)
	}
	host := gc.Request.Host
	params.Email = strings.ToLower(params.Email)

	if !utils.IsValidEmail(params.Email) {
		return res, fmt.Errorf("invalid email")
	}

	_, err = db.Provider.GetUserByEmail(params.Email)
	if err != nil {
		return res, fmt.Errorf(`user with this email not found`)
	}

	verificationToken, err := token.CreateVerificationToken(params.Email, constants.VerificationTypeForgotPassword)
	if err != nil {
		log.Println(`error generating token`, err)
	}
	db.Provider.AddVerificationRequest(models.VerificationRequest{
		Token:      verificationToken,
		Identifier: constants.VerificationTypeForgotPassword,
		ExpiresAt:  time.Now().Add(time.Minute * 30).Unix(),
		Email:      params.Email,
	})

	// exec it as go routin so that we can reduce the api latency
	go func() {
		email.SendForgotPasswordMail(params.Email, verificationToken, host)
	}()

	res = &model.Response{
		Message: `Please check your inbox! We have sent a password reset link.`,
	}

	return res, nil
}
