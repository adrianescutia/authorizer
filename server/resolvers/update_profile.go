package resolvers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/cookie"
	"github.com/authorizerdev/authorizer/server/db"
	"github.com/authorizerdev/authorizer/server/db/models"
	"github.com/authorizerdev/authorizer/server/email"
	"github.com/authorizerdev/authorizer/server/graph/model"
	"github.com/authorizerdev/authorizer/server/sessionstore"
	"github.com/authorizerdev/authorizer/server/token"
	"github.com/authorizerdev/authorizer/server/utils"
	"golang.org/x/crypto/bcrypt"
)

// UpdateProfileResolver is resolver for update profile mutation
func UpdateProfileResolver(ctx context.Context, params model.UpdateProfileInput) (*model.Response, error) {
	gc, err := utils.GinContextFromContext(ctx)
	var res *model.Response
	if err != nil {
		return res, err
	}

	claims, err := token.ValidateAccessToken(gc)
	if err != nil {
		return res, err
	}

	// validate if all params are not empty
	if params.GivenName == nil && params.FamilyName == nil && params.Picture == nil && params.MiddleName == nil && params.Nickname == nil && params.OldPassword == nil && params.Email == nil && params.Birthdate == nil && params.Gender == nil && params.PhoneNumber == nil {
		return res, fmt.Errorf("please enter at least one param to update")
	}

	userEmail := fmt.Sprintf("%v", claims["email"])
	user, err := db.Provider.GetUserByEmail(userEmail)
	if err != nil {
		return res, err
	}

	if params.GivenName != nil && user.GivenName != params.GivenName {
		user.GivenName = params.GivenName
	}

	if params.FamilyName != nil && user.FamilyName != params.FamilyName {
		user.FamilyName = params.FamilyName
	}

	if params.MiddleName != nil && user.MiddleName != params.MiddleName {
		user.MiddleName = params.MiddleName
	}

	if params.Nickname != nil && user.Nickname != params.Nickname {
		user.Nickname = params.Nickname
	}

	if params.Birthdate != nil && user.Birthdate != params.Birthdate {
		user.Birthdate = params.Birthdate
	}

	if params.Gender != nil && user.Gender != params.Gender {
		user.Gender = params.Gender
	}

	if params.PhoneNumber != nil && user.PhoneNumber != params.PhoneNumber {
		user.PhoneNumber = params.PhoneNumber
	}

	if params.Picture != nil && user.Picture != params.Picture {
		user.Picture = params.Picture
	}

	if params.OldPassword != nil {
		if err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(*params.OldPassword)); err != nil {
			return res, fmt.Errorf("incorrect old password")
		}

		if params.NewPassword == nil {
			return res, fmt.Errorf("new password is required")
		}

		if params.ConfirmNewPassword == nil {
			return res, fmt.Errorf("confirm password is required")
		}

		if *params.ConfirmNewPassword != *params.NewPassword {
			return res, fmt.Errorf(`password and confirm password does not match`)
		}

		password, _ := utils.EncryptPassword(*params.NewPassword)

		user.Password = &password
	}

	hasEmailChanged := false

	if params.Email != nil && user.Email != *params.Email {
		// check if valid email
		if !utils.IsValidEmail(*params.Email) {
			return res, fmt.Errorf("invalid email address")
		}
		newEmail := strings.ToLower(*params.Email)
		// check if user with new email exists
		_, err := db.Provider.GetUserByEmail(newEmail)

		// err = nil means user exists
		if err == nil {
			return res, fmt.Errorf("user with this email address already exists")
		}

		sessionstore.DeleteAllUserSession(fmt.Sprintf("%v", user.ID))
		cookie.DeleteCookie(gc)

		user.Email = newEmail
		user.EmailVerifiedAt = nil
		hasEmailChanged = true
		// insert verification request
		verificationType := constants.VerificationTypeUpdateEmail
		verificationToken, err := token.CreateVerificationToken(newEmail, verificationType)
		if err != nil {
			log.Println(`error generating token`, err)
		}
		db.Provider.AddVerificationRequest(models.VerificationRequest{
			Token:      verificationToken,
			Identifier: verificationType,
			ExpiresAt:  time.Now().Add(time.Minute * 30).Unix(),
			Email:      newEmail,
		})

		// exec it as go routin so that we can reduce the api latency
		go func() {
			email.SendVerificationMail(newEmail, verificationToken)
		}()
	}

	_, err = db.Provider.UpdateUser(user)
	if err != nil {
		log.Println("error updating user:", err)
		return res, err
	}
	message := `Profile details updated successfully.`
	if hasEmailChanged {
		message += `For the email change we have sent new verification email, please verify and continue`
	}
	res = &model.Response{
		Message: message,
	}

	return res, nil
}
